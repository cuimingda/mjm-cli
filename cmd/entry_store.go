package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type EntryStore struct {
	db *sql.DB
}

func NewEntryStore(path string) (*EntryStore, error) {
	if err := ensureDatabaseDirectory(path); err != nil {
		return nil, err
	}

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

func (s *EntryStore) List() ([]ScrapedEntry, error) {
	rows, err := s.db.Query(`SELECT href, title FROM scraped_entries ORDER BY href`)
	if err != nil {
		return nil, fmt.Errorf("query entries: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	entries := make([]ScrapedEntry, 0)
	for rows.Next() {
		var entry ScrapedEntry
		if err := rows.Scan(&entry.Href, &entry.Title); err != nil {
			return nil, fmt.Errorf("scan entry: %w", err)
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entries: %w", err)
	}

	return entries, nil
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

func ensureDatabaseDirectory(path string) error {
	directory := filepath.Dir(path)
	if directory == "." || directory == "" {
		return nil
	}

	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create database directory %q: %w", directory, err)
	}

	return nil
}
