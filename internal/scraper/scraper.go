package scraper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Scrape fetches the rendered HTML for a URL via scrape.do.
func Scrape(ctx context.Context, client *http.Client, targetURL string) (string, error) {
	if client == nil {
		client = http.DefaultClient
	}

	token := strings.TrimSpace(os.Getenv("SCRAPE_DO_TOKEN"))
	if token == "" {
		token = strings.TrimSpace(os.Getenv("SCRAPE_DO_API_KEY"))
	}
	if token == "" {
		return "", errors.New("SCRAPE_DO_TOKEN or SCRAPE_DO_API_KEY must be set")
	}

	if _, err := url.ParseRequestURI(targetURL); err != nil {
		return "", errors.New("invalid target URL")
	}

	encodedURL := url.QueryEscape(targetURL)
	reqURL := fmt.Sprintf("https://api.scrape.do/?url=%s&token=%s", encodedURL, token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("scrape request failed (%s): %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return string(body), nil
}
