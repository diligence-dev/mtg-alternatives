package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diligence-dev/mtg-alternatives/server"
)

func TestCardsPagination_Defaults(t *testing.T) {
	srv := newTestServer(t)

	for i := 0; i < 35; i++ {
		server.InsertAlternative(srv.DB(), "Card "+string(rune('A'+i)), "file.png")
	}

	req := httptest.NewRequest("GET", "/api/cards", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Cards []server.CardEntry `json:"cards"`
		Total int                `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Cards) != 30 {
		t.Fatalf("expected 30 cards (default page size), got %d", len(body.Cards))
	}
	if body.Total != 35 {
		t.Fatalf("expected total 35, got %d", body.Total)
	}
}

func TestCardsPagination_SecondPage(t *testing.T) {
	srv := newTestServer(t)

	for i := 0; i < 35; i++ {
		server.InsertAlternative(srv.DB(), "Card "+string(rune('A'+i)), "file.png")
	}

	req := httptest.NewRequest("GET", "/api/cards?page=2&limit=30", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Cards []server.CardEntry `json:"cards"`
		Total int                `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Cards) != 5 {
		t.Fatalf("expected 5 cards (page 2), got %d", len(body.Cards))
	}
	if body.Total != 35 {
		t.Fatalf("expected total 35, got %d", body.Total)
	}
}

func TestCardsPagination_CustomLimit(t *testing.T) {
	srv := newTestServer(t)

	for i := 0; i < 10; i++ {
		server.InsertAlternative(srv.DB(), "Card "+string(rune('A'+i)), "file.png")
	}

	req := httptest.NewRequest("GET", "/api/cards?limit=5", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Cards []server.CardEntry `json:"cards"`
		Total int                `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Cards) != 5 {
		t.Fatalf("expected 5 cards, got %d", len(body.Cards))
	}
	if body.Total != 10 {
		t.Fatalf("expected total 10, got %d", body.Total)
	}
}

func TestCardsPagination_EmptyDatabase(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest("GET", "/api/cards", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Cards []server.CardEntry `json:"cards"`
		Total int                `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Cards) != 0 {
		t.Fatalf("expected 0 cards, got %d", len(body.Cards))
	}
	if body.Total != 0 {
		t.Fatalf("expected total 0, got %d", body.Total)
	}
}
