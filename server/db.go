package server

import (
	"database/sql"
	"time"
	_ "github.com/mattn/go-sqlite3"
)

type Alternative struct {
	ID          int       `json:"id"`
	ScryfallID  string    `json:"scryfall_id"`
	Filename    string    `json:"filename"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS alternatives (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			scryfall_id TEXT NOT NULL,
			filename   TEXT NOT NULL,
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
		err := rows.Scan(&alt.ID, &alt.ScryfallID, &alt.Filename, &alt.UploadedAt)
		if err != nil {
			return nil, err
		}
		alternatives = append(alternatives, alt)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return alternatives, nil
}

func InsertAlternative(db *sql.DB, scryfallID, filename string) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO alternatives (scryfall_id, filename)
		VALUES (?, ?)
	`, scryfallID, filename)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}