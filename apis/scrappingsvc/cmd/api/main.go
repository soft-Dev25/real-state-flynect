package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	repositories "github.com/soft-Dev25/real-state-flynect/api/scrappingsvc/internal/core/repository"
	services "github.com/soft-Dev25/real-state-flynect/api/scrappingsvc/internal/core/service"
	"github.com/soft-Dev25/real-state-flynect/api/scrappingsvc/internal/handlers/http"
)

func main() {
	// Inicia el almacenamiento en memoria
	repo := repositories.NewMemoryRepo()

	// Crea el servicio de scraping
	scraperService := services.NewScraperService(repo)

	// Crea el handler HTTP
	scraperHandler := http.NewScraperHandler(scraperService)

	// Inicia Fiber
	app := fiber.New()

	// Usa el handler que ya hiciste
	app.Get("/scrape", func(c *fiber.Ctx) error {
		ctx := context.Background()
		err := scraperHandler.ScrapeFiber(ctx, c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return nil
	})

	log.Println("Scrapping service corriendo en http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}
