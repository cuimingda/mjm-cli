package cmd

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

type ScrapedEntry struct {
	Href  string
	Title string
}

type PageScraper struct{}

func NewPageScraper() *PageScraper {
	return &PageScraper{}
}

func (s *PageScraper) Scrape(pageURL string) ([]ScrapedEntry, error) {
	collector := colly.NewCollector()
	entries := make([]ScrapedEntry, 0)
	var scrapeErr error

	collector.OnHTML(".entry-title > a", func(element *colly.HTMLElement) {
		href := strings.TrimSpace(element.Request.AbsoluteURL(element.Attr("href")))
		title := strings.TrimSpace(element.Text)
		if href == "" || title == "" {
			return
		}

		entries = append(entries, ScrapedEntry{
			Href:  href,
			Title: title,
		})
	})

	collector.OnError(func(response *colly.Response, err error) {
		if response == nil || response.Request == nil || response.Request.URL == nil {
			scrapeErr = fmt.Errorf("scrape page: %w", err)
			return
		}

		scrapeErr = fmt.Errorf("scrape %s: %w", response.Request.URL.String(), err)
	})

	if err := collector.Visit(pageURL); err != nil {
		return nil, fmt.Errorf("visit page: %w", err)
	}

	if scrapeErr != nil {
		return nil, scrapeErr
	}

	return entries, nil
}
