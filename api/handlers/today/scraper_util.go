package today

import (
	"context"

	"github.com/user/Noiseless-service/api/handlers"
	"github.com/user/Noiseless-service/internal/scraper"
)

// ScrapeWebsite proxies the shared HTTP client into the scraper helper so other
// today components can fetch arbitrary pages when needed.
func ScrapeWebsite(ctx context.Context, targetURL string) (string, error) {
	return scraper.Scrape(ctx, handlers.HTTPClient, targetURL)
}
