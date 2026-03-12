package cmd

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

func TestDBTablesCommandDisplaysTableNamesAndRowCounts(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	paths := &AppPaths{
		developerID: defaultDeveloperID,
		appName:     defaultAppName,
		userConfigDir: func() (string, error) {
			return configDir, nil
		},
	}

	dbPath := filepath.Join(configDir, defaultDeveloperID, defaultAppName, defaultSQLiteFileName)
	createDatabaseForTablesTest(t, dbPath)

	output := executeCommand(t, newDBCommandWithPaths(paths), "tables")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 table lines, got %d: %q", len(lines), output)
	}

	if lines[0] != "episodes\t1" {
		t.Fatalf("unexpected first line %q", lines[0])
	}

	if lines[1] != "scraped_entries\t2" {
		t.Fatalf("unexpected second line %q", lines[1])
	}
}

func createDatabaseForTablesTest(t *testing.T, dbPath string) {
	t.Helper()

	store, err := NewEntryStore(dbPath)
	if err != nil {
		t.Fatalf("create entry store: %v", err)
	}

	if _, err := store.Save(ScrapedEntry{
		Href:  "https://www.meijumi.net/1.html",
		Title: "entry one",
	}); err != nil {
		t.Fatalf("save first entry: %v", err)
	}

	if _, err := store.Save(ScrapedEntry{
		Href:  "https://www.meijumi.net/2.html",
		Title: "entry two",
	}); err != nil {
		t.Fatalf("save second entry: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("close entry store: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if _, err := db.Exec(`
		CREATE TABLE episodes (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`); err != nil {
		t.Fatalf("create episodes table: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO episodes (name) VALUES ('episode 1')`); err != nil {
		t.Fatalf("insert episode row: %v", err)
	}
}
