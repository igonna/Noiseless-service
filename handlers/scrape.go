package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/Noiseless-service/models"
)

func (h *Handlers) Scrape(c *fiber.Ctx) error {
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

	body, err := h.deps.ScrapeWebsite(ctx, target)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	return c.JSON(models.ScrapeResponse{Body: body})
}
