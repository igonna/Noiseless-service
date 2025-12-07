package main

import (
	"context"
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

func handleProfileWebsites(c *fiber.Ctx) error {
	if _, err := authenticateSupabaseRequest(c); err != nil {
		return err
	}

	var req profileWebsitesRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON payload")
	}

	profile := strings.TrimSpace(req.Profile)
	if profile == "" {
		return fiber.NewError(fiber.StatusBadRequest, "profile is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	result, err := fetchProfileWebsites(ctx, profile)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	return c.JSON(result)
}
