package server_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/diligence-dev/mtg-alternatives/server"
)

func TestInitDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := server.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}
	defer db.Close()

	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='alternatives'").Scan(&name)
	if err != nil {
		t.Fatal("alternatives table not created")
	}
	if name != "alternatives" {
		t.Fatalf("expected table 'alternatives', got %q", name)
	}
}

func TestInitDB_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := server.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}
	db.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}
}

func TestGetAlternatives_Empty(t *testing.T) {
	db := openTestDB(t)

	results, err := server.GetAlternatives(db, "nonexistent-id")
	if err != nil {
		t.Fatalf("GetAlternatives returned error: %v", err)
	}
	if results != nil {
		t.Fatalf("expected nil for nonexistent ID, got %v", results)
	}
}

func TestInsertAndGetAlternatives(t *testing.T) {
	db := openTestDB(t)

	scryfallID := "test-card-123"
	filename := "abc-123.png"

	id, err := server.InsertAlternative(db, scryfallID, filename)
	if err != nil {
		t.Fatalf("InsertAlternative returned error: %v", err)
	}
	if id != 1 {
		t.Fatalf("expected id 1, got %d", id)
	}

	results, err := server.GetAlternatives(db, scryfallID)
	if err != nil {
		t.Fatalf("GetAlternatives returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ScryfallID != scryfallID {
		t.Errorf("expected scryfall_id %q, got %q", scryfallID, results[0].ScryfallID)
	}
	if results[0].Filename != filename {
		t.Errorf("expected filename %q, got %q", filename, results[0].Filename)
	}
}

func TestGetAlternatives_MultipleForSameCard(t *testing.T) {
	db := openTestDB(t)

	scryfallID := "test-card-123"
	server.InsertAlternative(db, scryfallID, "file1.png")
	server.InsertAlternative(db, scryfallID, "file2.png")

	results, err := server.GetAlternatives(db, scryfallID)
	if err != nil {
		t.Fatalf("GetAlternatives returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestGetAlternatives_OnlyReturnsMatchingID(t *testing.T) {
	db := openTestDB(t)

	server.InsertAlternative(db, "card-a", "file1.png")
	server.InsertAlternative(db, "card-b", "file2.png")

	results, err := server.GetAlternatives(db, "card-a")
	if err != nil {
		t.Fatalf("GetAlternatives returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ScryfallID != "card-a" {
		t.Errorf("expected scryfall_id 'card-a', got %q", results[0].ScryfallID)
	}
}