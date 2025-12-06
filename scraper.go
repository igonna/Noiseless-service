package main

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

func scrapeWebsite(ctx context.Context, targetURL string) (string, error) {
	token := strings.TrimSpace(os.Getenv("SCRAPE_DO_API_KEY"))
	if token == "" {
		return "", errors.New("SCRAPE_DO_API_KEY is not set")
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

	resp, err := httpClient.Do(req)
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
