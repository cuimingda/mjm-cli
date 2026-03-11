package cmd

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type EntryStore struct {
	db *sql.DB
}

func NewEntryStore(path string) (*EntryStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database %q: %w", path, err)
	}

	store := &EntryStore{db: db}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *EntryStore) Close() error {
	return s.db.Close()
}

func (s *EntryStore) Save(entry ScrapedEntry) (bool, error) {
	result, err := s.db.Exec(
		`INSERT OR IGNORE INTO scraped_entries (href, title) VALUES (?, ?)`,
		entry.Href,
		entry.Title,
	)
	if err != nil {
		return false, fmt.Errorf("insert entry %q: %w", entry.Href, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read inserted row count: %w", err)
	}

	return rowsAffected > 0, nil
}

func (s *EntryStore) init() error {
	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS scraped_entries (
			href TEXT PRIMARY KEY,
			title TEXT NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create scraped_entries table: %w", err)
	}

	return nil
}
