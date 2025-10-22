package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	services "github.com/soft-Dev25/real-state-flynect/api/scrappingsvc/internal/core/service"
)

// ScraperHandler representa un handler HTTP que usa el servicio de scraping
type ScraperHandler struct {
	scraperService *services.ScraperService
}

func NewScraperHandler(scraperService *services.ScraperService) *ScraperHandler {
	return &ScraperHandler{scraperService: scraperService}
}

func (h *ScraperHandler) HandleScrape(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Println("[ScraperHandler] Nueva solicitud de scraping recibida (HTTP).")
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	err := h.scraperService.Run(ctx)
	if err != nil {
		log.Printf("[ScraperHandler] Error al ejecutar scraping: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("[ScraperHandler] Scraping completado. Recuperando listados...")
	listings, err := h.scraperService.GetListings(ctx)
	if err != nil {
		log.Printf("[ScraperHandler] Error al obtener listados: %v\n", err)
		http.Error(w, "failed to fetch listings", http.StatusInternalServerError)
		return
	}
	log.Printf("[ScraperHandler] Se encontraron %d listados.\n", len(listings))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(listings); err != nil {
		log.Printf("[ScraperHandler] Error al codificar respuesta JSON: %v\n", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("[ScraperHandler] Solicitud completada en %v.\n", time.Since(start))
}

// ScrapeFiber permite usar el handler con Fiber
func (h *ScraperHandler) ScrapeFiber(ctx context.Context, c *fiber.Ctx) error {
	start := time.Now()
	log.Println("[ScraperHandler] Nueva solicitud de scraping recibida (Fiber).")
	log.Println("[ScraperHandler] Ejecutando proceso de scraping...")
	err := h.scraperService.Run(ctx)
	if err != nil {
		log.Printf("[ScraperHandler] Error durante el scraping: %v\n", err)
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	log.Println("[ScraperHandler] Scraping completado. Obteniendo listados...")
	listings, err := h.scraperService.GetListings(ctx)
	if err != nil {
		log.Printf("[ScraperHandler] Error al obtener listados: %v\n", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch listings")
	}
	log.Printf("[ScraperHandler] Se encontraron %d listados. Respondiendo al cliente...\n", len(listings))
	log.Printf("[ScraperHandler] Solicitud completada en %v.\n", time.Since(start))

	return c.JSON(listings)
}
