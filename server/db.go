package server

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Alternative struct {
	ID         int       `json:"id"`
	ScryfallID string    `json:"scryfall_id"`
	Filename   string    `json:"filename"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS alternatives (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			scryfall_id TEXT NOT NULL,
			filename    TEXT NOT NULL,
			uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_scryfall_id ON alternatives(scryfall_id);
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetAlternatives(db *sql.DB, scryfallID string) ([]Alternative, error) {
	rows, err := db.Query(`
		SELECT id, scryfall_id, filename, uploaded_at
		FROM alternatives
		WHERE scryfall_id = ?
		ORDER BY uploaded_at DESC
	`, scryfallID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alternatives []Alternative
	for rows.Next() {
		var alt Alternative
		if err := rows.Scan(&alt.ID, &alt.ScryfallID, &alt.Filename, &alt.UploadedAt); err != nil {
			return nil, err
		}
		alternatives = append(alternatives, alt)
	}

	return alternatives, rows.Err()
}

func InsertAlternative(db *sql.DB, scryfallID, filename string) (Alternative, error) {
	now := time.Now().UTC().Truncate(time.Second)
	result, err := db.Exec(`
		INSERT INTO alternatives (scryfall_id, filename, uploaded_at)
		VALUES (?, ?, ?)
	`, scryfallID, filename, now)
	if err != nil {
		return Alternative{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Alternative{}, err
	}

	return Alternative{
		ID:         int(id),
		ScryfallID: scryfallID,
		Filename:   filename,
		UploadedAt: now,
	}, nil
}
