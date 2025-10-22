package ports

import "github.com/soft-Dev25/real-state-flynect/api/scrappingsvc/internal/core/domain"

type ScraperPort interface {
	Scrape() ([]domain.Listing, error)
}
