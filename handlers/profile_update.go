package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) ProfileUpdate(c *fiber.Ctx) error {
	userID, err := h.deps.UserIDProvider(c)
	if err != nil {
		return err
	}

	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON payload")
	}

	if len(payload) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "profile payload cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	record, err := h.deps.UpsertProfile(ctx, userID, payload)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("profile upsert failed: %v", err))
	}

	return c.JSON(fiber.Map{"profile": record})
}
