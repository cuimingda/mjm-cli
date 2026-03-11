package cmd

import (
	"fmt"
	"io"
)

type ScrapeSummary struct {
	InsertedCount int
	SkippedCount  int
}

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

	entries, err := r.scrapePage(pageURL)
	if err != nil {
		return err
	}

	summary, err := r.storeEntries(store, entries)
	if err != nil {
		return err
	}

	if err := r.writeSummary(summary); err != nil {
		return err
	}

	return nil
}

func (r *ScrapeRunner) scrapePage(pageURL string) ([]ScrapedEntry, error) {
	return r.scraper.Scrape(pageURL)
}

func (r *ScrapeRunner) storeEntries(store *EntryStore, entries []ScrapedEntry) (ScrapeSummary, error) {
	summary := ScrapeSummary{}

	for _, entry := range entries {
		inserted, err := store.Save(entry)
		if err != nil {
			return ScrapeSummary{}, err
		}

		if inserted {
			summary.InsertedCount++
			continue
		}

		summary.SkippedCount++
		if _, err := fmt.Fprintf(r.stdout, "skip existing: %s\n", entry.Href); err != nil {
			return ScrapeSummary{}, fmt.Errorf("write duplicate output: %w", err)
		}
	}

	return summary, nil
}

func (r *ScrapeRunner) writeSummary(summary ScrapeSummary) error {
	if _, err := fmt.Fprintf(r.stdout, "completed: inserted=%d skipped=%d\n", summary.InsertedCount, summary.SkippedCount); err != nil {
		return fmt.Errorf("write summary output: %w", err)
	}

	return nil
}
