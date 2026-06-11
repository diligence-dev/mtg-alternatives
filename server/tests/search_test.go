package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diligence-dev/mtg-alternatives/server"
)

func TestSearch_EmptyQuery(t *testing.T) {
	srv := newTestServer(t, "")

	req := httptest.NewRequest("GET", "/api/search?q=", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var body server.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error != "Query parameter 'q' is required" {
		t.Fatalf("unexpected error: %q", body.Error)
	}
}

func TestSearch_MissingQueryParam(t *testing.T) {
	srv := newTestServer(t, "")

	req := httptest.NewRequest("GET", "/api/search", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSearch_PostMethod(t *testing.T) {
	srv := newTestServer(t, "")

	req := httptest.NewRequest("POST", "/api/search?q=test", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestSearch_Success(t *testing.T) {
	mock := mockScryfallServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("User-Agent header not set")
		}
		if r.Header.Get("Accept") == "" {
			t.Error("Accept header not set")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(server.ScryfallResponse{
			Data: []server.ScryfallCard{
				{
					ID:   "card-1",
					Name: "Test Card",
					ImageUris: struct {
						Normal string `json:"normal"`
					}{Normal: "https://example.com/card.png"},
				},
			},
		})
	})

	req := httptest.NewRequest("GET", "/api/search?q=test", nil)
	w := httptest.NewRecorder()
	mock.srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Cards []server.Card `json:"cards"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(body.Cards))
	}
	if body.Cards[0].ID != "card-1" {
		t.Errorf("expected card id 'card-1', got %q", body.Cards[0].ID)
	}
	if body.Cards[0].Name != "Test Card" {
		t.Errorf("expected card name 'Test Card', got %q", body.Cards[0].Name)
	}
	if body.Cards[0].Image != "https://example.com/card.png" {
		t.Errorf("unexpected image URL: %q", body.Cards[0].Image)
	}
}

func TestSearch_SkipsCardsWithoutImages(t *testing.T) {
	mock := mockScryfallServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(server.ScryfallResponse{
			Data: []server.ScryfallCard{
				{
					ID:   "card-1",
					Name: "Card With Image",
					ImageUris: struct {
						Normal string `json:"normal"`
					}{Normal: "https://example.com/card.png"},
				},
				{
					ID:   "card-2",
					Name: "Card Without Image",
					ImageUris: struct {
						Normal string `json:"normal"`
					}{Normal: ""},
				},
			},
		})
	})

	req := httptest.NewRequest("GET", "/api/search?q=test", nil)
	w := httptest.NewRecorder()
	mock.srv.ServeHTTP(w, req)

	var body struct {
		Cards []server.Card `json:"cards"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Cards) != 1 {
		t.Fatalf("expected 1 card (skip no-image), got %d", len(body.Cards))
	}
	if body.Cards[0].ID != "card-1" {
		t.Errorf("expected card-1, got %q", body.Cards[0].ID)
	}
}

func TestSearch_ScryfallError(t *testing.T) {
	mock := mockScryfallServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"details": "No cards found matching your query.",
		})
	})

	req := httptest.NewRequest("GET", "/api/search?q=zzzznonexistent", nil)
	w := httptest.NewRecorder()
	mock.srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}

	var body server.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error != "No cards found matching your query." {
		t.Fatalf("unexpected error: %q", body.Error)
	}
}

func TestSearch_ScryfallUnreachable(t *testing.T) {
	srv := newTestServer(t, "http://127.0.0.1:0")

	req := httptest.NewRequest("GET", "/api/search?q=test", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

type mockServer struct {
	srv *server.Server
}

func mockScryfallServer(t *testing.T, handler http.HandlerFunc) *mockServer {
	t.Helper()
	scryfall := httptest.NewServer(handler)
	t.Cleanup(scryfall.Close)
	srv := newTestServer(t, scryfall.URL)
	return &mockServer{srv: srv}
}
