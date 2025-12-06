package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func handleProfileUpdate(c *fiber.Ctx) error {
	user, err := authenticateSupabaseRequest(c)
	if err != nil {
		return err
	}

	var req profileUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON payload")
	}

	updates := map[string]interface{}{
		"id":         user.ID,
		"updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	if user.Email != "" {
		updates["email"] = user.Email
	}

	provided := false
	if req.DisplayName != nil {
		updates["display_name"] = strings.TrimSpace(*req.DisplayName)
		provided = true
	}
	if req.Bio != nil {
		updates["bio"] = strings.TrimSpace(*req.Bio)
		provided = true
	}
	if !provided {
		return fiber.NewError(fiber.StatusBadRequest, "provide at least one field to update")
	}

	client, err := getSupabaseClient()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var saved map[string]interface{}
	_, err = client.From("profiles").
		Upsert(updates, "id", "representation", "").
		Single().
		ExecuteTo(&saved)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, fmt.Sprintf("supabase upsert failed: %v", err))
	}

	return c.JSON(fiber.Map{"profile": saved})
}
