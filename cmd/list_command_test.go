package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestListCommandDisplaysAllEntries(t *testing.T) {
	t.Parallel()

	server := newSampleServer(t)
	dbPath := filepath.Join(t.TempDir(), "entries.sqlite")

	executeCommand(t, newScrapeCommand(), "--db", dbPath, server.URL)

	output := executeCommand(t, newListCommand(), "--db", dbPath)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 15 {
		t.Fatalf("expected 15 lines, got %d: %q", len(lines), output)
	}

	if lines[0] != "https://www.meijumi.net/31028.html\t《急诊室的故事(英版) 第三十五至三十六季》Casualty 迅雷下载" {
		t.Fatalf("unexpected first line %q", lines[0])
	}

	if lines[len(lines)-1] != "https://www.meijumi.net/44700.html\t《阿加莎·克里斯蒂之七面钟》Agatha Christie’s Seven Dials 迅雷下载" {
		t.Fatalf("unexpected last line %q", lines[len(lines)-1])
	}
}
