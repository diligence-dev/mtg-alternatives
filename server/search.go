package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
)

type ScryfallCard struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ImageUris struct {
		Normal string `json:"normal"`
	} `json:"image_uris"`
}

type ScryfallResponse struct {
	Data []ScryfallCard `json:"data"`
}

type Card struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		sendJSONError(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	req, err := http.NewRequest("GET", "https://api.scryfall.com/cards/search?q="+url.QueryEscape(query), nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		sendJSONError(w, "Failed to search Scryfall", http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", "MTGAlternatives/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := s.scryfallClient.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		sendJSONError(w, "Failed to search Scryfall", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read body: %v", err)
		sendJSONError(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		var scryfallErr struct {
			Details string `json:"details"`
		}
		json.Unmarshal(body, &scryfallErr)
		msg := scryfallErr.Details
		if msg == "" {
			msg = "Search failed"
		}
		log.Printf("Search failed with status %d: %s", resp.StatusCode, msg)
		sendJSONError(w, msg, resp.StatusCode)
		return
	}

	var scryfallResp ScryfallResponse
	if err := json.Unmarshal(body, &scryfallResp); err != nil {
		log.Printf("Failed to parse response: %v", err)
		sendJSONError(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}

	cards := make([]Card, 0, len(scryfallResp.Data))
	for _, c := range scryfallResp.Data {
		if c.ImageUris.Normal != "" {
			cards = append(cards, Card{ID: c.ID, Name: c.Name, Image: c.ImageUris.Normal})
		}
	}

	json.NewEncoder(w).Encode(map[string][]Card{"cards": cards})
}
