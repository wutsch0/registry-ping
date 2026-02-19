package registry

import "fmt"

// ScraperRegistry holds a list of scrapers and selects the right one for a given image.
type ScraperRegistry struct {
	scrapers []Scraper
}

// NewScraperRegistry creates a ScraperRegistry with the given scrapers.
func NewScraperRegistry(scrapers ...Scraper) *ScraperRegistry {
	return &ScraperRegistry{scrapers: scrapers}
}

// For returns the first scraper that can handle the image's registry host.
func (r *ScraperRegistry) For(ref ImageRef) (Scraper, error) {
	for _, s := range r.scrapers {
		if s.CanHandle(ref.Host) {
			return s, nil
		}
	}
	return nil, fmt.Errorf("no scraper registered for host %q", ref.Host)
}
