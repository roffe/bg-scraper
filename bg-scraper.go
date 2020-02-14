package bgscraper

// Scraper our scraper interface
type Scraper interface {
	Scrape() error
}
