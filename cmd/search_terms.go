package cmd

import (
	"fmt"
	"strings"
)

func normalizeSearchTerms(terms []string) ([]string, error) {
	normalized := make([]string, 0, len(terms))
	seen := make(map[string]struct{}, len(terms))

	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}

		key := strings.ToLower(term)
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}
		normalized = append(normalized, term)
	}

	if len(normalized) == 0 {
		return nil, fmt.Errorf("search requires at least one non-empty term")
	}

	return normalized, nil
}
