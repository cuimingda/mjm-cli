package cmd

import (
	"context"
	"fmt"
	"io"
	"sync"
)

type BatchRunner struct {
	pageURLBuilder *PageURLBuilder
	scrapeRunner   *ScrapeRunner
	storeFactory   func(path string) (*EntryStore, error)
	stdout         io.Writer
	outputMu       sync.Mutex
}

type batchPageRequest struct {
	pageNumber int
	pageURL    string
}

type batchPageResult struct {
	pageNumber int
	pageURL    string
	entries    []ScrapedEntry
	err        error
}

func NewBatchRunner(stdout io.Writer) *BatchRunner {
	if stdout == nil {
		stdout = io.Discard
	}

	return &BatchRunner{
		pageURLBuilder: NewPageURLBuilder(),
		scrapeRunner:   NewScrapeRunner(stdout),
		storeFactory:   NewEntryStore,
		stdout:         stdout,
	}
}

func (r *BatchRunner) Run(baseURL string, dbPath string, fromPage int, toPage int, parallelism int) error {
	pageRequests, err := r.buildPageRequests(baseURL, fromPage, toPage)
	if err != nil {
		return err
	}

	if parallelism > 0 {
		return r.runParallel(dbPath, pageRequests, parallelism)
	}

	return r.runSerial(dbPath, pageRequests)
}

func (r *BatchRunner) runSerial(dbPath string, pageRequests []batchPageRequest) error {
	store, err := r.storeFactory(dbPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = store.Close()
	}()

	for _, pageRequest := range pageRequests {
		if err := r.writeStart(pageRequest.pageURL); err != nil {
			return err
		}

		entries, err := r.scrapeRunner.scrapePage(pageRequest.pageURL)
		if err != nil {
			return err
		}

		if _, err := r.storePage(store, pageRequest.pageURL, entries); err != nil {
			return err
		}
	}

	return nil
}

func (r *BatchRunner) runParallel(dbPath string, pageRequests []batchPageRequest, parallelism int) error {
	store, err := r.storeFactory(dbPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = store.Close()
	}()

	results := r.scrapePagesInParallel(pageRequests, parallelism)
	var firstErr error

	for result := range results {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
			}
			continue
		}

		if firstErr != nil {
			continue
		}

		if _, err := r.storePage(store, result.pageURL, result.entries); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (r *BatchRunner) buildPageRequests(baseURL string, fromPage int, toPage int) ([]batchPageRequest, error) {
	pageRequests := make([]batchPageRequest, 0, toPage-fromPage+1)

	for pageNumber := fromPage; pageNumber <= toPage; pageNumber++ {
		pageURL, err := r.pageURLBuilder.Build(baseURL, pageNumber)
		if err != nil {
			return nil, err
		}

		pageRequests = append(pageRequests, batchPageRequest{
			pageNumber: pageNumber,
			pageURL:    pageURL,
		})
	}

	return pageRequests, nil
}

func (r *BatchRunner) storePage(store *EntryStore, pageURL string, entries []ScrapedEntry) (ScrapeSummary, error) {
	if len(entries) == 0 {
		return ScrapeSummary{}, fmt.Errorf("scrape %s: no entries matched .entry-title > a", pageURL)
	}

	summary, err := r.scrapeRunner.storeEntriesSilently(store, entries)
	if err != nil {
		return ScrapeSummary{}, err
	}

	if err := r.writeDone(pageURL, summary); err != nil {
		return ScrapeSummary{}, err
	}

	return summary, nil
}

func (r *BatchRunner) scrapePagesInParallel(pageRequests []batchPageRequest, parallelism int) <-chan batchPageResult {
	results := make(chan batchPageResult, len(pageRequests))
	jobs := make(chan batchPageRequest)
	ctx, cancel := context.WithCancel(context.Background())

	workerCount := parallelism
	if workerCount > len(pageRequests) {
		workerCount = len(pageRequests)
	}

	var workerGroup sync.WaitGroup
	for workerIndex := 0; workerIndex < workerCount; workerIndex++ {
		workerGroup.Add(1)
		go func() {
			defer workerGroup.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case pageRequest, ok := <-jobs:
					if !ok {
						return
					}

					if err := r.writeStart(pageRequest.pageURL); err != nil {
						results <- batchPageResult{
							pageNumber: pageRequest.pageNumber,
							pageURL:    pageRequest.pageURL,
							err:        err,
						}
						cancel()
						return
					}

					entries, err := r.scrapeRunner.scrapePage(pageRequest.pageURL)
					if err != nil {
						results <- batchPageResult{
							pageNumber: pageRequest.pageNumber,
							pageURL:    pageRequest.pageURL,
							err:        err,
						}
						cancel()
						return
					}

					if len(entries) == 0 {
						results <- batchPageResult{
							pageNumber: pageRequest.pageNumber,
							pageURL:    pageRequest.pageURL,
							err:        fmt.Errorf("scrape %s: no entries matched .entry-title > a", pageRequest.pageURL),
						}
						cancel()
						return
					}

					results <- batchPageResult{
						pageNumber: pageRequest.pageNumber,
						pageURL:    pageRequest.pageURL,
						entries:    entries,
					}
				}
			}
		}()
	}

	go func() {
		defer close(jobs)

		for _, pageRequest := range pageRequests {
			select {
			case <-ctx.Done():
				return
			case jobs <- pageRequest:
			}
		}
	}()

	go func() {
		workerGroup.Wait()
		cancel()
		close(results)
	}()

	return results
}

func (r *BatchRunner) writeStart(pageURL string) error {
	r.outputMu.Lock()
	defer r.outputMu.Unlock()

	if _, err := fmt.Fprintf(r.stdout, "start: url=%s\n", pageURL); err != nil {
		return fmt.Errorf("write batch start output: %w", err)
	}

	return nil
}

func (r *BatchRunner) writeDone(pageURL string, summary ScrapeSummary) error {
	r.outputMu.Lock()
	defer r.outputMu.Unlock()

	if _, err := fmt.Fprintf(
		r.stdout,
		"done: url=%s found=%d saved=%d skipped=%d\n",
		pageURL,
		summary.FoundCount,
		summary.InsertedCount,
		summary.SkippedCount,
	); err != nil {
		return fmt.Errorf("write batch done output: %w", err)
	}

	return nil
}
