package cmd

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestBatchCommandScrapesMultiplePages(t *testing.T) {
	t.Parallel()

	sampleHTML := mustReadSampleHTML(t)
	pageTwoHTML := strings.ReplaceAll(sampleHTML, "https://www.meijumi.net/", "https://www.meijumi.net/page2/")
	server := newBatchServer(t, map[string]batchResponse{
		"/en/page/1/": {
			statusCode: http.StatusOK,
			body:       sampleHTML,
		},
		"/en/page/2/": {
			statusCode: http.StatusOK,
			body:       pageTwoHTML,
		},
	})

	dbPath := filepath.Join(t.TempDir(), "batch.sqlite")
	output := executeCommand(t, newBatchCommand(), "--db", dbPath, "--to", "2", server.URL+"/en/")

	if strings.Count(output, "completed: inserted=15 skipped=0") != 2 {
		t.Fatalf("expected two completed summaries, got %q", output)
	}

	entryCount, err := countEntries(t, dbPath)
	if err != nil {
		t.Fatalf("count entries: %v", err)
	}

	if entryCount != 30 {
		t.Fatalf("expected 30 entries, got %d", entryCount)
	}

	title, err := lookupTitle(t, dbPath, "https://www.meijumi.net/page2/44700.html")
	if err != nil {
		t.Fatalf("lookup page 2 title: %v", err)
	}

	if title != "《阿加莎·克里斯蒂之七面钟》Agatha Christie’s Seven Dials 迅雷下载" {
		t.Fatalf("unexpected page 2 title %q", title)
	}
}

func TestBatchCommandValidatesPageRange(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		args       []string
		wantSubstr string
	}{
		{
			name:       "missing to",
			args:       []string{"https://example.com/en/"},
			wantSubstr: `required flag(s) "to" not set`,
		},
		{
			name:       "from less than one",
			args:       []string{"--from", "0", "--to", "1", "https://example.com/en/"},
			wantSubstr: "--from must be a positive integer greater than or equal to 1",
		},
		{
			name:       "to less than one",
			args:       []string{"--to", "0", "https://example.com/en/"},
			wantSubstr: "--to must be a positive integer greater than or equal to 1",
		},
		{
			name:       "to smaller than from",
			args:       []string{"--from", "3", "--to", "2", "https://example.com/en/"},
			wantSubstr: "--to must be greater than or equal to --from",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := executeCommandErr(newBatchCommand(), testCase.args...)
			if err == nil {
				t.Fatal("expected command error")
			}

			if !strings.Contains(err.Error(), testCase.wantSubstr) {
				t.Fatalf("expected error containing %q, got %q", testCase.wantSubstr, err.Error())
			}
		})
	}
}

func TestBatchCommandStopsOnNotFoundPage(t *testing.T) {
	t.Parallel()

	sampleHTML := mustReadSampleHTML(t)
	server := newBatchServer(t, map[string]batchResponse{
		"/en/page/1/": {
			statusCode: http.StatusOK,
			body:       sampleHTML,
		},
		"/en/page/2/": {
			statusCode: http.StatusNotFound,
			body:       "not found",
		},
	})

	dbPath := filepath.Join(t.TempDir(), "batch-404.sqlite")
	output, err := executeCommandErr(newBatchCommand(), "--db", dbPath, "--to", "3", server.URL+"/en/")
	if err == nil {
		t.Fatal("expected command error")
	}

	if !strings.Contains(err.Error(), server.URL+"/en/page/2/") {
		t.Fatalf("expected error to mention page 2 url, got %q", err.Error())
	}

	if !strings.Contains(output, "completed: inserted=15 skipped=0") {
		t.Fatalf("expected page 1 summary before failure, got %q", output)
	}

	entryCount, err := countEntries(t, dbPath)
	if err != nil {
		t.Fatalf("count entries: %v", err)
	}

	if entryCount != 15 {
		t.Fatalf("expected only first page data to be stored, got %d", entryCount)
	}
}

func TestBatchCommandStopsOnPageWithoutEntries(t *testing.T) {
	t.Parallel()

	sampleHTML := mustReadSampleHTML(t)
	server := newBatchServer(t, map[string]batchResponse{
		"/en/page/1/": {
			statusCode: http.StatusOK,
			body:       sampleHTML,
		},
		"/en/page/2/": {
			statusCode: http.StatusOK,
			body:       "<html><body>empty page</body></html>",
		},
	})

	dbPath := filepath.Join(t.TempDir(), "batch-empty.sqlite")
	output, err := executeCommandErr(newBatchCommand(), "--db", dbPath, "--to", "2", server.URL+"/en/")
	if err == nil {
		t.Fatal("expected command error")
	}

	if !strings.Contains(err.Error(), "no entries matched .entry-title > a") {
		t.Fatalf("expected empty-page error, got %q", err.Error())
	}

	if !strings.Contains(output, "completed: inserted=15 skipped=0") {
		t.Fatalf("expected page 1 summary before failure, got %q", output)
	}

	entryCount, err := countEntries(t, dbPath)
	if err != nil {
		t.Fatalf("count entries: %v", err)
	}

	if entryCount != 15 {
		t.Fatalf("expected only first page data to be stored, got %d", entryCount)
	}
}

func TestPageURLBuilderBuildsPaginatedURL(t *testing.T) {
	t.Parallel()

	builder := NewPageURLBuilder()
	pageURL, err := builder.Build("https://www.meijumi.net/en/", 2)
	if err != nil {
		t.Fatalf("build page url: %v", err)
	}

	if pageURL != "https://www.meijumi.net/en/page/2/" {
		t.Fatalf("unexpected page url %q", pageURL)
	}
}

func TestPageURLBuilderRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	builder := NewPageURLBuilder()
	testCases := []struct {
		name    string
		baseURL string
		page    int
	}{
		{
			name:    "invalid url",
			baseURL: "://bad-url",
			page:    1,
		},
		{
			name:    "page less than one",
			baseURL: "https://www.meijumi.net/en/",
			page:    0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := builder.Build(testCase.baseURL, testCase.page)
			if err == nil {
				t.Fatal("expected builder error")
			}
		})
	}
}

type batchResponse struct {
	statusCode int
	body       string
}

func newBatchServer(t *testing.T, responses map[string]batchResponse) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		response, ok := responses[request.URL.Path]
		if !ok {
			http.NotFound(writer, request)
			return
		}

		writer.WriteHeader(response.statusCode)
		_, _ = writer.Write([]byte(response.body))
	}))

	t.Cleanup(server.Close)

	return server
}

func mustReadSampleHTML(t *testing.T) string {
	t.Helper()

	content, err := readSampleHTML()
	if err != nil {
		t.Fatalf("read sample html: %v", err)
	}

	return content
}
