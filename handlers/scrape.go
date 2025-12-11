package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/Noiseless-service/models"
)

func Scrape(c *fiber.Ctx) error {
	var req models.ScrapeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON payload")
	}

	target := strings.TrimSpace(req.URL)
	if target == "" {
		return fiber.NewError(fiber.StatusBadRequest, "url is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	body, err := scrapeWebsite(ctx, target)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	return c.JSON(models.ScrapeResponse{Body: body})
}

func scrapeWebsite(ctx context.Context, targetURL string) (string, error) {
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

	client := HTTPClient
	if client == nil {
		client = http.DefaultClient
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
