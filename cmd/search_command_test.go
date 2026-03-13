package cmd

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSearchCommandDisplaysMatchesForMultipleTerms(t *testing.T) {
	t.Parallel()

	server := newSampleServer(t)
	dbPath := filepath.Join(t.TempDir(), "entries.sqlite")

	executeCommand(t, newScrapeCommand(), "--db", dbPath, server.URL)

	output := executeCommand(t, newSearchCommand(), "--db", dbPath, "Seven", "Dials")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %q", len(lines), output)
	}

	expected := "https://www.meijumi.net/44700.html\t《阿加莎·克里斯蒂之七面钟》Agatha Christie’s " +
		"\x1b[1;38;5;196mSeven\x1b[0m " +
		"\x1b[1;38;5;208mDials\x1b[0m 迅雷下载"
	if lines[0] != expected {
		t.Fatalf("unexpected output line %q", lines[0])
	}
}

func TestSearchCommandMatchesShortFragments(t *testing.T) {
	t.Parallel()

	server := newSampleServer(t)
	dbPath := filepath.Join(t.TempDir(), "entries.sqlite")

	executeCommand(t, newScrapeCommand(), "--db", dbPath, server.URL)

	output := executeCommand(t, newSearchCommand(), "--db", dbPath, "七面")
	line := strings.TrimSpace(output)

	expected := "https://www.meijumi.net/44700.html\t《阿加莎·克里斯蒂之" +
		"\x1b[1;38;5;196m七面\x1b[0m钟》Agatha Christie’s Seven Dials 迅雷下载"
	if line != expected {
		t.Fatalf("unexpected output line %q", line)
	}
}

func TestSearchCommandRebuildsIndexForLegacyDatabase(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "legacy.sqlite")
	createLegacyScrapedEntriesDB(t, dbPath, []ScrapedEntry{
		{
			Href:  "https://example.com/legacy",
			Title: "Legacy Girl Taken Sample",
		},
	})

	output := executeCommand(t, newSearchCommand(), "--db", dbPath, "Girl", "Taken")
	line := strings.TrimSpace(output)

	expected := "https://example.com/legacy\tLegacy " +
		"\x1b[1;38;5;196mGirl\x1b[0m " +
		"\x1b[1;38;5;208mTaken\x1b[0m Sample"
	if line != expected {
		t.Fatalf("unexpected output line %q", line)
	}
}

func TestSearchCommandSortsByHrefByDefault(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "sorting-default.sqlite")
	createLegacyScrapedEntriesDB(t, dbPath, []ScrapedEntry{
		{
			Href:  "https://example.com/z-entry",
			Title: "Alpha Sample",
		},
		{
			Href:  "https://example.com/a-entry",
			Title: "Zulu Sample",
		},
	})

	output := executeCommand(t, newSearchCommand(), "--db", dbPath, "Sample")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), output)
	}

	expected := []string{
		"https://example.com/a-entry\tZulu \x1b[1;38;5;196mSample\x1b[0m",
		"https://example.com/z-entry\tAlpha \x1b[1;38;5;196mSample\x1b[0m",
	}
	if strings.Join(lines, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected output lines %q", lines)
	}
}

func TestSearchCommandSortsByTitleWhenFlagIsSet(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "sorting-title.sqlite")
	createLegacyScrapedEntriesDB(t, dbPath, []ScrapedEntry{
		{
			Href:  "https://example.com/z-entry",
			Title: "Alpha Sample",
		},
		{
			Href:  "https://example.com/a-entry",
			Title: "Zulu Sample",
		},
	})

	output := executeCommand(t, newSearchCommand(), "--db", dbPath, "--sort-by-title", "Sample")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), output)
	}

	expected := []string{
		"https://example.com/z-entry\tAlpha \x1b[1;38;5;196mSample\x1b[0m",
		"https://example.com/a-entry\tZulu \x1b[1;38;5;196mSample\x1b[0m",
	}
	if strings.Join(lines, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected output lines %q", lines)
	}
}

func createLegacyScrapedEntriesDB(t *testing.T, dbPath string, entries []ScrapedEntry) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if _, err := db.Exec(`
		CREATE TABLE scraped_entries (
			href TEXT PRIMARY KEY,
			title TEXT NOT NULL
		)
	`); err != nil {
		t.Fatalf("create legacy scraped_entries table: %v", err)
	}

	for _, entry := range entries {
		if _, err := db.Exec(`INSERT INTO scraped_entries (href, title) VALUES (?, ?)`, entry.Href, entry.Title); err != nil {
			t.Fatalf("insert legacy entry %q: %v", entry.Href, err)
		}
	}
}
