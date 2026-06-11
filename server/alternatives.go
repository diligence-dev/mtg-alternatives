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

type AlternativeResponse struct {
	ID         int    `json:"id"`
	URL        string `json:"url"`
	UploadedAt string `json:"uploaded_at"`
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

	alts, err := GetAlternatives(s.db, scryfallID)
	if err != nil {
		log.Printf("Failed to fetch alternatives: %v", err)
		sendJSONError(w, "Failed to fetch alternatives", http.StatusInternalServerError)
		return
	}

	resp := make([]AlternativeResponse, len(alts))
	for i, alt := range alts {
		resp[i] = AlternativeResponse{
			ID:         alt.ID,
			URL:        fmt.Sprintf("/uploads/%s", alt.Filename),
			UploadedAt: alt.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	json.NewEncoder(w).Encode(map[string][]AlternativeResponse{"alternatives": resp})
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

	allowed := map[string]bool{"image/png": true, "image/jpeg": true, "image/webp": true}
	if !allowed[header.Header.Get("Content-Type")] {
		sendJSONError(w, "Invalid file type. Only PNG, JPEG, and WebP are allowed", http.StatusBadRequest)
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".png"
	}
	filename := uuid.New().String() + ext
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

	alt, err := InsertAlternative(s.db, scryfallID, filename)
	if err != nil {
		os.Remove(dstPath)
		log.Printf("Failed to insert alternative: %v", err)
		sendJSONError(w, "Failed to save alternative", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]AlternativeResponse{
		"alternative": {
			ID:         alt.ID,
			URL:        fmt.Sprintf("/uploads/%s", alt.Filename),
			UploadedAt: alt.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}
