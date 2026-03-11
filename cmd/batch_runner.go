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
	return &BatchRunner{
		pageURLBuilder: NewPageURLBuilder(),
		scrapeRunner:   NewScrapeRunner(stdout),
		storeFactory:   NewEntryStore,
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
		entries, err := r.scrapeRunner.scrapePage(pageRequest.pageURL)
		if err != nil {
			return err
		}

		if err := r.storePage(store, pageRequest.pageURL, entries); err != nil {
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
	successes := make(map[int]batchPageResult, len(pageRequests))
	var firstError *batchPageResult

	for result := range results {
		if result.err != nil {
			if firstError == nil || result.pageNumber < firstError.pageNumber {
				resultCopy := result
				firstError = &resultCopy
			}
			continue
		}

		successes[result.pageNumber] = result
	}

	lastSuccessfulPage := pageRequests[len(pageRequests)-1].pageNumber
	if firstError != nil {
		lastSuccessfulPage = firstError.pageNumber - 1
	}

	for _, pageRequest := range pageRequests {
		if pageRequest.pageNumber > lastSuccessfulPage {
			break
		}

		result, ok := successes[pageRequest.pageNumber]
		if !ok {
			return fmt.Errorf("scrape %s: missing result before batch termination", pageRequest.pageURL)
		}

		if err := r.storePage(store, result.pageURL, result.entries); err != nil {
			return err
		}
	}

	if firstError != nil {
		return firstError.err
	}

	return nil
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

func (r *BatchRunner) storePage(store *EntryStore, pageURL string, entries []ScrapedEntry) error {
	if len(entries) == 0 {
		return fmt.Errorf("scrape %s: no entries matched .entry-title > a", pageURL)
	}

	summary, err := r.scrapeRunner.storeEntries(store, entries)
	if err != nil {
		return err
	}

	if err := r.scrapeRunner.writeSummary(summary); err != nil {
		return err
	}

	return nil
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
		defer cancel()

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
		close(results)
	}()

	return results
}
