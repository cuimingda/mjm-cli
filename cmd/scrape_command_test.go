package cmd

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScrapeCommandStoresEntriesAndSkipsDuplicates(t *testing.T) {
	t.Parallel()

	server := newSampleServer(t)
	dbPath := filepath.Join(t.TempDir(), "entries.sqlite")

	firstOutput := executeScrapeCommand(t, server.URL, dbPath)
	if !strings.Contains(firstOutput, "completed: inserted=15 skipped=0") {
		t.Fatalf("expected first run summary, got %q", firstOutput)
	}

	entryCount, err := countEntries(t, dbPath)
	if err != nil {
		t.Fatalf("count entries: %v", err)
	}

	if entryCount != 15 {
		t.Fatalf("expected 15 entries, got %d", entryCount)
	}

	title, err := lookupTitle(t, dbPath, "https://www.meijumi.net/44672.html")
	if err != nil {
		t.Fatalf("lookup title: %v", err)
	}

	if title != "《失窃的女孩第一季》Girl Taken 迅雷下载" {
		t.Fatalf("unexpected title %q", title)
	}

	secondOutput := executeScrapeCommand(t, server.URL, dbPath)
	if !strings.Contains(secondOutput, "skip existing: https://www.meijumi.net/44672.html") {
		t.Fatalf("expected duplicate skip output, got %q", secondOutput)
	}

	if !strings.Contains(secondOutput, "completed: inserted=0 skipped=15") {
		t.Fatalf("expected second run summary, got %q", secondOutput)
	}
}

func newSampleServer(t *testing.T) *httptest.Server {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("..", "data", "sample.html"))
	if err != nil {
		t.Fatalf("read sample html: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(content)
	}))

	t.Cleanup(server.Close)

	return server
}

func executeScrapeCommand(t *testing.T, pageURL string, dbPath string) string {
	t.Helper()

	command := newScrapeCommand()
	output := &bytes.Buffer{}
	command.SetOut(output)
	command.SetErr(output)
	command.SetArgs([]string{"--db", dbPath, pageURL})

	if err := command.Execute(); err != nil {
		t.Fatalf("execute scrape command: %v", err)
	}

	return output.String()
}

func countEntries(t *testing.T, dbPath string) (int, error) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = db.Close()
	}()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM scraped_entries`).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func lookupTitle(t *testing.T, dbPath string, href string) (string, error) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = db.Close()
	}()

	var title string
	if err := db.QueryRow(`SELECT title FROM scraped_entries WHERE href = ?`, href).Scan(&title); err != nil {
		return "", err
	}

	return title, nil
}
