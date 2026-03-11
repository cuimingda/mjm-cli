package cmd

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type PageURLBuilder struct{}

func NewPageURLBuilder() *PageURLBuilder {
	return &PageURLBuilder{}
}

func (b *PageURLBuilder) Build(baseURL string, pageNumber int) (string, error) {
	if pageNumber < 1 {
		return "", fmt.Errorf("page number must be greater than or equal to 1")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base url %q: %w", baseURL, err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", fmt.Errorf("invalid base url %q", baseURL)
	}

	parsedURL.Path = ensureTrailingSlash(path.Join(parsedURL.Path, "page", strconv.Itoa(pageNumber)))
	parsedURL.RawPath = ""
	parsedURL.Fragment = ""

	return parsedURL.String(), nil
}

func ensureTrailingSlash(value string) string {
	if strings.HasSuffix(value, "/") {
		return value
	}

	return value + "/"
}
