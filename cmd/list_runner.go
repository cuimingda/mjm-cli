package cmd

import (
	"fmt"
	"io"
)

type ListRunner struct {
	storeFactory func(path string) (*EntryStore, error)
	stdout       io.Writer
}

func NewListRunner(stdout io.Writer) *ListRunner {
	if stdout == nil {
		stdout = io.Discard
	}

	return &ListRunner{
		storeFactory: NewEntryStore,
		stdout:       stdout,
	}
}

func (r *ListRunner) Run(dbPath string) error {
	store, err := r.storeFactory(dbPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = store.Close()
	}()

	entries, err := store.List()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if _, err := fmt.Fprintf(r.stdout, "%s\t%s\n", entry.Href, entry.Title); err != nil {
			return fmt.Errorf("write entry output: %w", err)
		}
	}

	return nil
}
