package server_test

import (
	"database/sql"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/diligence-dev/mtg-alternatives/server"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := server.InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func newTestServer(t *testing.T, scryfallURL string) *server.Server {
	t.Helper()

	db := openTestDB(t)
	uploadsDir := t.TempDir()

	var client *http.Client
	if scryfallURL != "" {
		client = &http.Client{
			Transport: &scryfallURLRewriter{baseURL: scryfallURL},
		}
	} else {
		client = &http.Client{
			Timeout: 5 * 1000000000,
		}
	}

	return server.NewServerWithClient(db, uploadsDir, nil, client)
}

type scryfallURLRewriter struct {
	baseURL string
}

func (r *scryfallURLRewriter) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq, _ := http.NewRequest(req.Method, r.baseURL+req.URL.Path+"?"+req.URL.RawQuery, req.Body)
	newReq.Header = req.Header.Clone()
	return http.DefaultTransport.RoundTrip(newReq)
}