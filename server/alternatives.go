package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type AlternativesResponse struct {
	Alternatives []AlternativeResponse `json:"alternatives"`
	Error        string               `json:"error,omitempty"`
}

type AlternativeResponse struct {
	ID         int    `json:"id"`
	URL        string `json:"url"`
	UploadedAt string `json:"uploaded_at"`
}

type UploadResponse struct {
	Alternative AlternativeResponse `json:"alternative"`
	Error        string             `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (s *Server) handleAlternatives(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		s.getAlternatives(w, r)
	case http.MethodPost:
		s.uploadAlternative(w, r)
	default:
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getAlternatives(w http.ResponseWriter, r *http.Request) {
	scryfallID := r.URL.Query().Get("scryfall_id")
	if scryfallID == "" {
		sendJSONError(w, "scryfall_id parameter is required", http.StatusBadRequest)
		return
	}

	alternatives, err := GetAlternatives(s.db, scryfallID)
	if err != nil {
		log.Printf("Failed to fetch alternatives: %v", err)
		sendJSONError(w, "Failed to fetch alternatives", http.StatusInternalServerError)
		return
	}

	response := AlternativesResponse{
		Alternatives: make([]AlternativeResponse, len(alternatives)),
	}

	for i, alt := range alternatives {
		response.Alternatives[i] = AlternativeResponse{
			ID:         alt.ID,
			URL:        fmt.Sprintf("/uploads/%s", alt.Filename),
			UploadedAt: alt.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) uploadAlternative(w http.ResponseWriter, r *http.Request) {
	// Limit upload size to 5 MB
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)
	if err := r.ParseMultipartForm(5 * 1024 * 1024); err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		sendJSONError(w, "File too large (max 5MB)", http.StatusBadRequest)
		return
	}

	scryfallID := r.FormValue("scryfall_id")
	if scryfallID == "" {
		sendJSONError(w, "scryfall_id is required", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("Failed to get form file: %v", err)
		sendJSONError(w, "image file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	allowedTypes := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/webp": true,
	}
	if !allowedTypes[header.Header.Get("Content-Type")] {
		sendJSONError(w, "Invalid file type. Only PNG, JPEG, and WebP are allowed", http.StatusBadRequest)
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		// Default to .png if no extension
		ext = ".png"
	}
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Save file to uploads directory
	dstPath := filepath.Join(s.uploadsDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		sendJSONError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Failed to copy file: %v", err)
		sendJSONError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Insert into database
	id, err := InsertAlternative(s.db, scryfallID, filename)
	if err != nil {
		// Clean up file if database insert fails
		os.Remove(dstPath)
		log.Printf("Failed to insert alternative: %v", err)
		sendJSONError(w, "Failed to save alternative", http.StatusInternalServerError)
		return
	}

	// Fetch the inserted alternative to get uploaded_at
	alternatives, err := GetAlternatives(s.db, scryfallID)
	if err != nil {
		log.Printf("Failed to fetch alternatives after insert: %v", err)
		sendJSONError(w, "Failed to fetch alternative", http.StatusInternalServerError)
		return
	}

	var uploadedAlt Alternative
	for _, alt := range alternatives {
		if int(id) == alt.ID {
			uploadedAlt = alt
			break
		}
	}

	response := UploadResponse{
		Alternative: AlternativeResponse{
			ID:         uploadedAlt.ID,
			URL:        fmt.Sprintf("/uploads/%s", uploadedAlt.Filename),
			UploadedAt: uploadedAlt.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}