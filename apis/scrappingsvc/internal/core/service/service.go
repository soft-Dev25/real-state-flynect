package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/soft-Dev25/real-state-flynect/api/scrappingsvc/internal/core/domain"
	"github.com/soft-Dev25/real-state-flynect/api/scrappingsvc/internal/core/ports"
)

type ScraperService struct {
	repo ports.StoragePort
}

func NewScraperService(repo ports.StoragePort) *ScraperService {
	return &ScraperService{repo: repo}
}

func (s *ScraperService) Run(ctx context.Context) error {
	start := time.Now()
	log.Println("[ScraperService] Inicio del scraping de sitios...")
	listings, err := s.scrapeSites(ctx)
	if err != nil {
		log.Printf("[ScraperService] Error al hacer scraping: %v\n", err)
		return fmt.Errorf("error scraping sites: %w", err)
	}
	log.Printf("[ScraperService] Se obtuvieron %d anuncios.\n", len(listings))
	for i, l := range listings {
		log.Printf("[ScraperService] Guardando anuncio %d: %s (%s)\n", i+1, l.Title, l.Location)
		if err := s.repo.SaveListing(ctx, l); err != nil {
			log.Printf("[ScraperService] Error al guardar el anuncio '%s': %v\n", l.Title, err)
			return err
		}
	}
	log.Printf("[ScraperService] Proceso completado en %v\n", time.Since(start))
	return nil
}

func (s *ScraperService) GetListings(ctx context.Context) ([]domain.Listing, error) {
	return s.repo.GetListings(ctx)
}

func (s *ScraperService) scrapeSites(ctx context.Context) ([]domain.Listing, error) {
	var listings []domain.Listing

	c := colly.NewCollector()
	c.OnHTML(".posting-card", func(e *colly.HTMLElement) {
		title := e.ChildText(".posting-card__title")
		price := e.ChildText(".price__value")
		location := e.ChildText(".posting-card__location")

		listings = append(listings, domain.Listing{
			Title:    title,
			PriceMXN: parsePrice(price),
			Location: location,
			Source:   "Inmuebles24",
		})
	})

	_ = c.Visit("https://www.inmuebles24.com/departamentos-en-renta.html")

	return listings, nil
}

func parsePrice(priceStr string) float64 {
	var price float64
	fmt.Sscanf(priceStr, "$%f", &price)
	return price
}
