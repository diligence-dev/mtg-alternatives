package server

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Alternative struct {
	ID         int       `json:"id"`
	ScryfallID string    `json:"scryfall_id"`
	Name       string    `json:"name"`
	Filename   string    `json:"filename"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type CardEntry struct {
	ScryfallID string `json:"scryfall_id"`
	Name       string `json:"name"`
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
			name        TEXT NOT NULL DEFAULT '',
			filename    TEXT NOT NULL,
			uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_scryfall_id ON alternatives(scryfall_id);
	`)
	if err != nil {
		return nil, err
	}

	migrateAddName(db)

	return db, nil
}

func migrateAddName(db *sql.DB) {
	row := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('alternatives') WHERE name='name'")
	var count int
	if err := row.Scan(&count); err != nil || count > 0 {
		return
	}
	db.Exec("ALTER TABLE alternatives ADD COLUMN name TEXT NOT NULL DEFAULT ''")
}

func GetAlternatives(db *sql.DB, scryfallID string) ([]Alternative, error) {
	rows, err := db.Query(`
		SELECT id, scryfall_id, name, filename, uploaded_at
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
		if err := rows.Scan(&alt.ID, &alt.ScryfallID, &alt.Name, &alt.Filename, &alt.UploadedAt); err != nil {
			return nil, err
		}
		alternatives = append(alternatives, alt)
	}

	return alternatives, rows.Err()
}

func InsertAlternative(db *sql.DB, scryfallID, name, filename string) (Alternative, error) {
	now := time.Now().UTC().Truncate(time.Second)
	result, err := db.Exec(`
		INSERT INTO alternatives (scryfall_id, name, filename, uploaded_at)
		VALUES (?, ?, ?, ?)
	`, scryfallID, name, filename, now)
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
		Name:       name,
		Filename:   filename,
		UploadedAt: now,
	}, nil
}

func GetCardsWithAlternatives(db *sql.DB) ([]CardEntry, error) {
	rows, err := db.Query(`
		SELECT DISTINCT scryfall_id, name
		FROM alternatives
		ORDER BY scryfall_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []CardEntry
	for rows.Next() {
		var c CardEntry
		if err := rows.Scan(&c.ScryfallID, &c.Name); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}

	return cards, rows.Err()
}
