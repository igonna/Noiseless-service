package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/Noiseless-service/models"
)

func (h *Handlers) TodaySignals(c *fiber.Ctx) error {
	userID, err := h.deps.UserIDProvider(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	websites, refreshed, err := h.deps.EnsureWebsitesSync(ctx, userID)
	if err != nil {
		if h.deps.ProfileMissingFunc(err) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"signals":            []models.Signal{},
		"websites":           websites.Websites,
		"websites_refreshed": refreshed,
	})
}
