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

	results, err := server.GetAlternatives(db, "Nonexistent Card")
	if err != nil {
		t.Fatalf("GetAlternatives returned error: %v", err)
	}
	if results != nil {
		t.Fatalf("expected nil for nonexistent name, got %v", results)
	}
}

func TestInsertAndGetAlternatives(t *testing.T) {
	db := openTestDB(t)

	alt, err := server.InsertAlternative(db, "Lightning Bolt", "abc-123.png")
	if err != nil {
		t.Fatalf("InsertAlternative returned error: %v", err)
	}
	if alt.ID != 1 {
		t.Fatalf("expected id 1, got %d", alt.ID)
	}
	if alt.Name != "Lightning Bolt" {
		t.Errorf("expected name 'Lightning Bolt', got %q", alt.Name)
	}
	if alt.Filename != "abc-123.png" {
		t.Errorf("expected filename 'abc-123.png', got %q", alt.Filename)
	}

	results, err := server.GetAlternatives(db, "Lightning Bolt")
	if err != nil {
		t.Fatalf("GetAlternatives returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Lightning Bolt" {
		t.Errorf("expected name 'Lightning Bolt', got %q", results[0].Name)
	}
	if results[0].Filename != "abc-123.png" {
		t.Errorf("expected filename 'abc-123.png', got %q", results[0].Filename)
	}
}

func TestGetAlternatives_MultipleForSameCard(t *testing.T) {
	db := openTestDB(t)

	server.InsertAlternative(db, "Lightning Bolt", "file1.png")
	server.InsertAlternative(db, "Lightning Bolt", "file2.png")

	results, err := server.GetAlternatives(db, "Lightning Bolt")
	if err != nil {
		t.Fatalf("GetAlternatives returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestGetAlternatives_OnlyReturnsMatchingName(t *testing.T) {
	db := openTestDB(t)

	server.InsertAlternative(db, "Card A", "file1.png")
	server.InsertAlternative(db, "Card B", "file2.png")

	results, err := server.GetAlternatives(db, "Card A")
	if err != nil {
		t.Fatalf("GetAlternatives returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Card A" {
		t.Errorf("expected name 'Card A', got %q", results[0].Name)
	}
}

func TestGetCardsWithAlternatives(t *testing.T) {
	db := openTestDB(t)

	server.InsertAlternative(db, "Card A", "file1.png")
	server.InsertAlternative(db, "Card A", "file2.png")
	server.InsertAlternative(db, "Card B", "file3.png")

	cards, err := server.GetCardsWithAlternatives(db)
	if err != nil {
		t.Fatalf("GetCardsWithAlternatives returned error: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}

	names := map[string]bool{}
	for _, c := range cards {
		names[c.Name] = true
	}
	if !names["Card A"] {
		t.Error("expected card 'Card A'")
	}
	if !names["Card B"] {
		t.Error("expected card 'Card B'")
	}
}
