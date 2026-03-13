package cmd

import "testing"

func TestTitleHighlighterHighlightsMultipleOccurrencesCaseInsensitively(t *testing.T) {
	t.Parallel()

	highlighter, err := NewTitleHighlighter([]string{"seven"})
	if err != nil {
		t.Fatalf("new title highlighter: %v", err)
	}

	got := highlighter.Highlight("Seven SEVEN seven")
	want := "\x1b[1;38;5;196mSeven\x1b[0m " +
		"\x1b[1;38;5;196mSEVEN\x1b[0m " +
		"\x1b[1;38;5;196mseven\x1b[0m"
	if got != want {
		t.Fatalf("unexpected highlighted title %q", got)
	}
}

func TestTitleHighlighterSkipsDuplicateSearchTerms(t *testing.T) {
	t.Parallel()

	highlighter, err := NewTitleHighlighter([]string{"Girl", "girl", " GIRL "})
	if err != nil {
		t.Fatalf("new title highlighter: %v", err)
	}

	got := highlighter.Highlight("Girl girl")
	want := "\x1b[1;38;5;196mGirl\x1b[0m " +
		"\x1b[1;38;5;196mgirl\x1b[0m"
	if got != want {
		t.Fatalf("unexpected highlighted title %q", got)
	}
}
