package server

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Alternative struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Filename   string    `json:"filename"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type CardEntry struct {
	Name string `json:"name"`
}

type CardsPage struct {
	Cards []CardEntry `json:"cards"`
	Total int         `json:"total"`
}

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS alternatives (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT NOT NULL,
			filename    TEXT NOT NULL,
			uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_name ON alternatives(name);
	`)
	if err != nil {
		return nil, err
	}

	migrateDropScryfallID(db)

	return db, nil
}

func migrateDropScryfallID(db *sql.DB) {
	row := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('alternatives') WHERE name='scryfall_id'")
	var count int
	if err := row.Scan(&count); err != nil || count == 0 {
		return
	}
	db.Exec(`CREATE TABLE alternatives_new (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		name        TEXT NOT NULL,
		filename    TEXT NOT NULL,
		uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	db.Exec(`INSERT INTO alternatives_new (id, name, filename, uploaded_at) SELECT id, name, filename, uploaded_at FROM alternatives`)
	db.Exec(`DROP TABLE alternatives`)
	db.Exec(`ALTER TABLE alternatives_new RENAME TO alternatives`)
	db.Exec(`CREATE INDEX idx_name ON alternatives(name)`)
}

func GetAlternatives(db *sql.DB, name string) ([]Alternative, error) {
	rows, err := db.Query(`
		SELECT id, name, filename, uploaded_at
		FROM alternatives
		WHERE name = ?
		ORDER BY uploaded_at DESC
	`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alternatives []Alternative
	for rows.Next() {
		var alt Alternative
		if err := rows.Scan(&alt.ID, &alt.Name, &alt.Filename, &alt.UploadedAt); err != nil {
			return nil, err
		}
		alternatives = append(alternatives, alt)
	}

	return alternatives, rows.Err()
}

func InsertAlternative(db *sql.DB, name, filename string) (Alternative, error) {
	now := time.Now().UTC().Truncate(time.Second)
	result, err := db.Exec(`
		INSERT INTO alternatives (name, filename, uploaded_at)
		VALUES (?, ?, ?)
	`, name, filename, now)
	if err != nil {
		return Alternative{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Alternative{}, err
	}

	return Alternative{
		ID:         int(id),
		Name:       name,
		Filename:   filename,
		UploadedAt: now,
	}, nil
}

func GetCardsWithAlternatives(db *sql.DB) ([]CardEntry, error) {
	rows, err := db.Query(`
		SELECT DISTINCT name
		FROM alternatives
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []CardEntry
	for rows.Next() {
		var c CardEntry
		if err := rows.Scan(&c.Name); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}

	return cards, rows.Err()
}

func GetCardsWithAlternativesPaginated(db *sql.DB, page, limit int) (CardsPage, error) {
	var total int
	err := db.QueryRow(`SELECT COUNT(DISTINCT name) FROM alternatives`).Scan(&total)
	if err != nil {
		return CardsPage{}, err
	}

	offset := (page - 1) * limit
	rows, err := db.Query(`
		SELECT DISTINCT name
		FROM alternatives
		ORDER BY uploaded_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return CardsPage{}, err
	}
	defer rows.Close()

	var cards []CardEntry
	for rows.Next() {
		var c CardEntry
		if err := rows.Scan(&c.Name); err != nil {
			return CardsPage{}, err
		}
		cards = append(cards, c)
	}

	if cards == nil {
		cards = []CardEntry{}
	}

	return CardsPage{Cards: cards, Total: total}, rows.Err()
}
