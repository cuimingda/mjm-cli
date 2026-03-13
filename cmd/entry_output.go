package cmd

import (
	"fmt"
	"io"
)

func writeEntries(stdout io.Writer, entries []ScrapedEntry) error {
	if stdout == nil {
		stdout = io.Discard
	}

	for _, entry := range entries {
		if _, err := fmt.Fprintf(stdout, "%s\t%s\n", entry.Href, entry.Title); err != nil {
			return fmt.Errorf("write entry output: %w", err)
		}
	}

	return nil
}

func writeHighlightedEntries(stdout io.Writer, entries []ScrapedEntry, highlighter *TitleHighlighter) error {
	if stdout == nil {
		stdout = io.Discard
	}

	for _, entry := range entries {
		title := entry.Title
		if highlighter != nil {
			title = highlighter.Highlight(entry.Title)
		}

		if _, err := fmt.Fprintf(stdout, "%s\t%s\n", entry.Href, title); err != nil {
			return fmt.Errorf("write entry output: %w", err)
		}
	}

	return nil
}
