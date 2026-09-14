package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// SQLiteStore persists URL mappings to a SQLite database file.
// The short code is derived from the row's auto-incrementing id (base62-encoded),
// so no separate counter needs to be tracked in memory.
type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS urls (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		long_url   TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating schema: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) Save(longURL string) (string, error) {
	result, err := s.db.Exec("INSERT INTO urls (long_url) VALUES (?)", longURL)
	if err != nil {
		return "", fmt.Errorf("inserting url: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("getting insert id: %w", err)
	}
	return encodeBase62(uint64(id)), nil
}

func (s *SQLiteStore) Get(code string) (string, bool, error) {
	id, err := decodeBase62(code)
	if err != nil {
		return "", false, nil
	}

	var longURL string
	err = s.db.QueryRow("SELECT long_url FROM urls WHERE id = ?", id).Scan(&longURL)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("querying url: %w", err)
	}
	return longURL, true, nil
}
