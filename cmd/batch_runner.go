package cmd

import (
	"fmt"
	"io"
)

type BatchRunner struct {
	pageURLBuilder *PageURLBuilder
	scrapeRunner   *ScrapeRunner
	storeFactory   func(path string) (*EntryStore, error)
}

func NewBatchRunner(stdout io.Writer) *BatchRunner {
	return &BatchRunner{
		pageURLBuilder: NewPageURLBuilder(),
		scrapeRunner:   NewScrapeRunner(stdout),
		storeFactory:   NewEntryStore,
	}
}

func (r *BatchRunner) Run(baseURL string, dbPath string, fromPage int, toPage int) error {
	store, err := r.storeFactory(dbPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = store.Close()
	}()

	for pageNumber := fromPage; pageNumber <= toPage; pageNumber++ {
		pageURL, err := r.pageURLBuilder.Build(baseURL, pageNumber)
		if err != nil {
			return err
		}

		entries, err := r.scrapeRunner.scrapePage(pageURL)
		if err != nil {
			return err
		}

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
	}

	return nil
}
