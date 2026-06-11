package server_test

import (
	"database/sql"
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

func newTestServer(t *testing.T) *server.Server {
	t.Helper()
	db := openTestDB(t)
	uploadsDir := t.TempDir()
	return server.NewServer(db, uploadsDir, nil)
}
