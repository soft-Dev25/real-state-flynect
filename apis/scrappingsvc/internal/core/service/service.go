package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
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
	log.Println("[ScraperService] 🚀 Iniciando scraping de Inmuebles24 (robusto y optimizado)...")

	listings, err := s.scrapeSites(ctx)
	if err != nil {
		log.Printf("[ScraperService] ❌ Error durante el scraping: %v\n", err)
		return fmt.Errorf("error scraping sites: %w", err)
	}

	log.Printf("[ScraperService] ✅ Se obtuvieron %d anuncios.\n", len(listings))

	for i, l := range listings {
		log.Printf("[ScraperService] Guardando anuncio %d: %s (%s)\n", i+1, l.Title, l.Location)
		if err := s.repo.SaveListing(ctx, l); err != nil {
			log.Printf("[ScraperService] ⚠️ Error al guardar el anuncio '%s': %v\n", l.Title, err)
		}
	}

	log.Printf("[ScraperService] 🏁 Scraping completado en %v\n", time.Since(start))
	return nil
}

func (s *ScraperService) GetListings(ctx context.Context) ([]domain.Listing, error) {
	return s.repo.GetListings(ctx)
}

func (s *ScraperService) scrapeSites(ctx context.Context) ([]domain.Listing, error) {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	// if _, err := os.Stat("inmuebles24_1761165243.html"); err == nil {
	// 	log.Println("[ScraperService] 🧠 Cargando desde archivo local en lugar de navegar (HTML real)...")
	// 	htmlBytes, err := os.ReadFile("inmuebles24_1761165243.html")
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error leyendo archivo local: %w", err)
	// 	}

	// 	log.Printf("[ScraperService] 📄 HTML cargado correctamente (%d bytes)", len(htmlBytes))

	// 	_ = os.WriteFile("inmuebles24_debug.html", htmlBytes, 0644)
	// 	log.Println("[ScraperService] 🧾 Copia del HTML analizado guardada en inmuebles24_debug.html")

	// 	snippet := string(htmlBytes)
	// 	if len(snippet) > 1000 {
	// 		snippet = snippet[:1000]
	// 	}
	// 	log.Println("[ScraperService] 🔍 Primeros 1000 caracteres del HTML:")
	// 	log.Println(snippet)

	// 	return parseInmuebles24FromHTML(string(htmlBytes))
	// }

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-translate", true),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64)
			AppleWebKit/537.36 (KHTML, like Gecko)
			Chrome/118.0.5993.88 Safari/537.36`),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	url := "https://www.inmuebles24.com/departamentos-en-renta-en-ciudad-de-mexico.html"
	log.Printf("[ScraperService] 🌐 Visitando página: %s", url)

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(10*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		log.Printf("[ScraperService] ❌ Error al cargar página: %v", err)
		return nil, err
	}

	_ = os.WriteFile("inmuebles24_snapshot.html", []byte(htmlContent), 0644)

	if strings.Contains(htmlContent, "Cloudflare") || strings.Contains(htmlContent, "blocked") {
		log.Println("[ScraperService] 🚫 Cloudflare detectado, reintentando con modo visible...")
		return s.scrapeWithVisibleBrowser(ctx, url)
	}

	log.Println("[ScraperService] 🧩 HTML renderizado guardado en inmuebles24_snapshot.html")

	listings, err := parseInmuebles24FromHTML(htmlContent)
	if err != nil {
		log.Printf("[ScraperService] ⚠️ Error analizando HTML: %v", err)
		return nil, err
	}

	log.Printf("[ScraperService] ✅ Se detectaron %d anuncios válidos.", len(listings))
	return listings, nil
}

// 🚀 Reintento con navegador visible (por si Cloudflare bloquea el headless)
func (s *ScraperService) scrapeWithVisibleBrowser(ctx context.Context, url string) ([]domain.Listing, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64)
			AppleWebKit/537.36 (KHTML, like Gecko)
			Chrome/118.0.5993.88 Safari/537.36`),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`[data-qa="POSTING_CARD_PRICE"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)

	if err != nil {
		return nil, fmt.Errorf("error al recargar página visible: %w", err)
	}

	_ = os.WriteFile("inmuebles24_snapshot_retry.html", []byte(htmlContent), 0644)
	log.Println("[ScraperService] ✅ Página guardada (modo visible). Analizando HTML...")

	return parseInmuebles24FromHTML(htmlContent)
}

// 🧩 Parser local desde HTML con JSON-LD + fallback regex
func parseInmuebles24FromHTML(html string) ([]domain.Listing, error) {
	// 1️⃣ Extraer bloques JSON-LD
	reJSON := regexp.MustCompile(`<script[^>]+type="application/ld\+json"[^>]*>(.*?)</script>`)
	jsonBlocks := reJSON.FindAllStringSubmatch(html, -1)

	var listings []domain.Listing

	for _, block := range jsonBlocks {
		raw := strings.TrimSpace(block[1])
		if !strings.Contains(raw, `"@type": "Apartment"`) {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &data); err != nil {
			continue
		}

		name := fmt.Sprint(data["name"])
		image := fmt.Sprint(data["image"])
		address := ""
		if addr, ok := data["address"].(map[string]interface{}); ok {
			address = fmt.Sprint(addr["name"])
		}

		listings = append(listings, domain.Listing{
			Title:    cleanHTML(name),
			PriceMXN: 0,
			Location: cleanHTML(address),
			Link:     image,
			Source:   "Inmuebles24",
		})
	}

	if len(listings) > 0 {
		log.Printf("[ScraperService] ✅ Extraídos %d anuncios desde JSON-LD.", len(listings))
		return listings, nil
	}

	// 2️⃣ Fallback: regex sobre HTML
	log.Println("[ScraperService] ⚠️ No se encontraron listados JSON-LD, aplicando fallback regex...")
	reTitle := regexp.MustCompile(`<h2[^>]*class="[^"]*postingCardTitle[^"]*"[^>]*>(.*?)</h2>`)
	rePrice := regexp.MustCompile(`<div[^>]*class="[^"]*postingCardPrice[^"]*"[^>]*>(.*?)</div>`)
	reLoc := regexp.MustCompile(`<span[^>]*class="[^"]*postingCardLocation[^"]*"[^>]*>(.*?)</span>`)
	reLink := regexp.MustCompile(`<a[^>]*class="[^"]*go-to-posting[^"]*"[^>]*href="(.*?)"`)

	titles := reTitle.FindAllStringSubmatch(html, -1)
	prices := rePrice.FindAllStringSubmatch(html, -1)
	locations := reLoc.FindAllStringSubmatch(html, -1)
	links := reLink.FindAllStringSubmatch(html, -1)

	for i := 0; i < len(titles) && i < len(prices); i++ {
		listings = append(listings, domain.Listing{
			Title:    cleanHTML(titles[i][1]),
			PriceMXN: parsePrice(cleanHTML(prices[i][1])),
			Location: func() string {
				if i < len(locations) {
					return cleanHTML(locations[i][1])
				}
				return ""
			}(),
			Link: func() string {
				if i < len(links) {
					return "https://www.inmuebles24.com" + cleanHTML(links[i][1])
				}
				return ""
			}(),
			Source: "Inmuebles24",
		})
	}

	log.Printf("[ScraperService] ✅ Extraídos %d anuncios desde HTML local.", len(listings))
	return listings, nil
}

func cleanHTML(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.TrimSpace(s)
	return s
}

func parsePrice(priceStr string) float64 {
	priceStr = strings.NewReplacer("$", "", ",", "", "MXN", "", "/mes", "", "Desde", "").Replace(priceStr)
	priceStr = strings.TrimSpace(priceStr)
	f, _ := strconv.ParseFloat(priceStr, 64)
	return f
}
