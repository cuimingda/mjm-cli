package cmd

import (
	"strings"
	"unicode"
)

const ansiReset = "\x1b[0m"

var searchHighlightColors = []string{
	"\x1b[1;38;5;196m",
	"\x1b[1;38;5;208m",
	"\x1b[1;38;5;220m",
	"\x1b[1;38;5;46m",
	"\x1b[1;38;5;51m",
	"\x1b[1;38;5;39m",
	"\x1b[1;38;5;93m",
	"\x1b[1;38;5;201m",
	"\x1b[1;38;5;154m",
	"\x1b[1;38;5;45m",
}

type TitleHighlighter struct {
	matchers []titleMatcher
}

type titleMatcher struct {
	lowercase []rune
	color     string
}

func NewTitleHighlighter(terms []string) (*TitleHighlighter, error) {
	normalizedTerms, err := normalizeSearchTerms(terms)
	if err != nil {
		return nil, err
	}

	matchers := make([]titleMatcher, 0, len(normalizedTerms))
	for index, term := range normalizedTerms {
		matchers = append(matchers, titleMatcher{
			lowercase: toLowerRunes([]rune(term)),
			color:     searchHighlightColors[index%len(searchHighlightColors)],
		})
	}

	return &TitleHighlighter{
		matchers: matchers,
	}, nil
}

func (h *TitleHighlighter) Highlight(title string) string {
	if h == nil || len(h.matchers) == 0 || title == "" {
		return title
	}

	titleRunes := []rune(title)
	lowercaseTitleRunes := toLowerRunes(titleRunes)
	owners := make([]int, len(titleRunes))
	for index := range owners {
		owners[index] = -1
	}

	for matcherIndex, matcher := range h.matchers {
		matcherLength := len(matcher.lowercase)
		if matcherLength == 0 || matcherLength > len(titleRunes) {
			continue
		}

		for start := 0; start <= len(titleRunes)-matcherLength; start++ {
			if !hasRunePrefix(lowercaseTitleRunes[start:], matcher.lowercase) {
				continue
			}

			for offset := 0; offset < matcherLength; offset++ {
				if owners[start+offset] != -1 {
					continue
				}

				owners[start+offset] = matcherIndex
			}
		}
	}

	var builder strings.Builder
	currentOwner := -1

	for index, titleRune := range titleRunes {
		if owners[index] != currentOwner {
			if currentOwner != -1 {
				builder.WriteString(ansiReset)
			}

			currentOwner = owners[index]
			if currentOwner != -1 {
				builder.WriteString(h.matchers[currentOwner].color)
			}
		}

		builder.WriteRune(titleRune)
	}

	if currentOwner != -1 {
		builder.WriteString(ansiReset)
	}

	return builder.String()
}

func hasRunePrefix(value []rune, prefix []rune) bool {
	if len(prefix) > len(value) {
		return false
	}

	for index := range prefix {
		if value[index] != prefix[index] {
			return false
		}
	}

	return true
}

func toLowerRunes(value []rune) []rune {
	lowercase := make([]rune, len(value))
	for index, item := range value {
		lowercase[index] = unicode.ToLower(item)
	}

	return lowercase
}
