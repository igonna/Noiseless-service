package main

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

const userIDContextKey = "user_id"

func userIDFromCtx(c *fiber.Ctx) (string, error) {
	raw := c.Locals(userIDContextKey)
	if raw == nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, "missing Supabase user context")
	}

	userID, _ := raw.(string)
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "missing Supabase user context")
	}

	return userID, nil
}
