package cmd

import (
	"fmt"
	"io"
)

type ScrapeRunner struct {
	scraper      *PageScraper
	storeFactory func(path string) (*EntryStore, error)
	stdout       io.Writer
}

func NewScrapeRunner(stdout io.Writer) *ScrapeRunner {
	if stdout == nil {
		stdout = io.Discard
	}

	return &ScrapeRunner{
		scraper:      NewPageScraper(),
		storeFactory: NewEntryStore,
		stdout:       stdout,
	}
}

func (r *ScrapeRunner) Run(pageURL string, dbPath string) error {
	store, err := r.storeFactory(dbPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = store.Close()
	}()

	entries, err := r.scraper.Scrape(pageURL)
	if err != nil {
		return err
	}

	insertedCount := 0
	skippedCount := 0

	for _, entry := range entries {
		inserted, err := store.Save(entry)
		if err != nil {
			return err
		}

		if inserted {
			insertedCount++
			continue
		}

		skippedCount++
		if _, err := fmt.Fprintf(r.stdout, "skip existing: %s\n", entry.Href); err != nil {
			return fmt.Errorf("write duplicate output: %w", err)
		}
	}

	if _, err := fmt.Fprintf(r.stdout, "completed: inserted=%d skipped=%d\n", insertedCount, skippedCount); err != nil {
		return fmt.Errorf("write summary output: %w", err)
	}

	return nil
}
