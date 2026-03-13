package cmd

import "io"

type SearchRunner struct {
	storeFactory func(path string) (*EntryStore, error)
	stdout       io.Writer
}

func NewSearchRunner(stdout io.Writer) *SearchRunner {
	if stdout == nil {
		stdout = io.Discard
	}

	return &SearchRunner{
		storeFactory: NewEntryStore,
		stdout:       stdout,
	}
}

func (r *SearchRunner) Run(dbPath string, terms []string) error {
	store, err := r.storeFactory(dbPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = store.Close()
	}()

	entries, err := store.Search(terms)
	if err != nil {
		return err
	}

	return writeEntries(r.stdout, entries)
}
