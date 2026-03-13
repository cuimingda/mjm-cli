package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	_ "modernc.org/sqlite"
)

type EntryStore struct {
	db *sql.DB
}

var titleSearchPatternEscaper = strings.NewReplacer(
	`\`, `\\`,
	`%`, `\%`,
	`_`, `\_`,
)

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

func (s *EntryStore) Search(fragments []string) ([]ScrapedEntry, error) {
	searchPlan, err := buildTitleSearchPlan(fragments)
	if err != nil {
		return nil, err
	}

	query := `SELECT href, title FROM scraped_entries`
	conditions := make([]string, 0, 1+len(searchPlan.likePatterns))
	args := make([]any, 0, 1+len(searchPlan.likePatterns))

	if searchPlan.matchQuery != "" {
		conditions = append(conditions, `rowid IN (SELECT rowid FROM scraped_entries_fts WHERE scraped_entries_fts MATCH ?)`)
		args = append(args, searchPlan.matchQuery)
	}

	for _, likePattern := range searchPlan.likePatterns {
		conditions = append(conditions, `title LIKE ? ESCAPE '\'`)
		args = append(args, likePattern)
	}

	query += ` WHERE ` + strings.Join(conditions, ` AND `) + ` ORDER BY href`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query matching entries: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	entries := make([]ScrapedEntry, 0)
	for rows.Next() {
		var entry ScrapedEntry
		if err := rows.Scan(&entry.Href, &entry.Title); err != nil {
			return nil, fmt.Errorf("scan matching entry: %w", err)
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate matching entries: %w", err)
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

	if _, err := s.db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS scraped_entries_fts
		USING fts5(
			title,
			content='scraped_entries',
			content_rowid='rowid',
			tokenize='trigram'
		)
	`); err != nil {
		return fmt.Errorf("create scraped_entries_fts table: %w", err)
	}

	if _, err := s.db.Exec(`
		CREATE TRIGGER IF NOT EXISTS scraped_entries_ai AFTER INSERT ON scraped_entries BEGIN
			INSERT INTO scraped_entries_fts(rowid, title)
			VALUES (new.rowid, new.title);
		END
	`); err != nil {
		return fmt.Errorf("create scraped_entries insert trigger: %w", err)
	}

	if _, err := s.db.Exec(`
		CREATE TRIGGER IF NOT EXISTS scraped_entries_ad AFTER DELETE ON scraped_entries BEGIN
			INSERT INTO scraped_entries_fts(scraped_entries_fts, rowid, title)
			VALUES ('delete', old.rowid, old.title);
		END
	`); err != nil {
		return fmt.Errorf("create scraped_entries delete trigger: %w", err)
	}

	if _, err := s.db.Exec(`
		CREATE TRIGGER IF NOT EXISTS scraped_entries_au AFTER UPDATE ON scraped_entries BEGIN
			INSERT INTO scraped_entries_fts(scraped_entries_fts, rowid, title)
			VALUES ('delete', old.rowid, old.title);
			INSERT INTO scraped_entries_fts(rowid, title)
			VALUES (new.rowid, new.title);
		END
	`); err != nil {
		return fmt.Errorf("create scraped_entries update trigger: %w", err)
	}

	if err := s.rebuildSearchIndex(); err != nil {
		return err
	}

	return nil
}

func (s *EntryStore) rebuildSearchIndex() error {
	if _, err := s.db.Exec(`INSERT INTO scraped_entries_fts(scraped_entries_fts) VALUES ('rebuild')`); err != nil {
		return fmt.Errorf("rebuild scraped_entries_fts: %w", err)
	}

	return nil
}

type titleSearchPlan struct {
	matchQuery   string
	likePatterns []string
}

func buildTitleSearchPlan(fragments []string) (titleSearchPlan, error) {
	normalizedTerms, err := normalizeSearchTerms(fragments)
	if err != nil {
		return titleSearchPlan{}, err
	}

	ftsTerms := make([]string, 0, len(normalizedTerms))
	likePatterns := make([]string, 0, len(normalizedTerms))

	for _, term := range normalizedTerms {
		if utf8.RuneCountInString(term) < 3 {
			likePatterns = append(likePatterns, "%"+escapeLikePattern(term)+"%")
			continue
		}

		ftsTerms = append(ftsTerms, `"`+escapeFTS5Phrase(term)+`"`)
	}

	return titleSearchPlan{
		matchQuery:   strings.Join(ftsTerms, ` AND `),
		likePatterns: likePatterns,
	}, nil
}

func escapeFTS5Phrase(value string) string {
	return strings.ReplaceAll(value, `"`, `""`)
}

func escapeLikePattern(value string) string {
	return titleSearchPatternEscaper.Replace(value)
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
