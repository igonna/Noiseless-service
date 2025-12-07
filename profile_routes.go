package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const profileUpsertSQL = `
INSERT INTO profiles (id, profile)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET
	profile = excluded.profile
RETURNING id, profile;
`

const profileSelectSQL = `
SELECT id, profile
FROM profiles
WHERE id = $1;
`

func handleProfileUpdate(c *fiber.Ctx) error {
	claims, err := supabaseClaimsFromCtx(c)
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

	record, err := upsertProfileRecord(ctx, claims.Subject, payload)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("profile upsert failed: %v", err))
	}

	return c.JSON(fiber.Map{"profile": record})
}

func handleProfileWebsites(c *fiber.Ctx) error {
	if _, err := supabaseClaimsFromCtx(c); err != nil {
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

func handleProfileGet(c *fiber.Ctx) error {
	claims, err := supabaseClaimsFromCtx(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	record, err := fetchProfileRecord(ctx, claims.Subject)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "profile not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"profile": record})
}

func supabaseClaimsFromCtx(c *fiber.Ctx) (*SupabaseClaims, error) {
	raw := c.Locals("claims")
	claims, ok := raw.(*SupabaseClaims)
	if !ok || claims == nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "missing Supabase claims")
	}
	if claims.Subject == "" {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "token missing subject")
	}
	return claims, nil
}

type profileRecord struct {
	ID      string                 `json:"id"`
	Profile map[string]interface{} `json:"profile"`
}

func upsertProfileRecord(ctx context.Context, userID string, profile map[string]interface{}) (profileRecord, error) {
	db, err := getDB()
	if err != nil {
		return profileRecord{}, err
	}

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return profileRecord{}, err
	}

	row := db.QueryRowContext(ctx, profileUpsertSQL, userID, profileJSON)

	var record profileRecord
	var raw json.RawMessage

	if err := row.Scan(&record.ID, &raw); err != nil {
		return profileRecord{}, err
	}

	if err := json.Unmarshal(raw, &record.Profile); err != nil {
		return profileRecord{}, err
	}

	return record, nil
}

func fetchProfileRecord(ctx context.Context, userID string) (profileRecord, error) {
	db, err := getDB()
	if err != nil {
		return profileRecord{}, err
	}

	row := db.QueryRowContext(ctx, profileSelectSQL, userID)
	var record profileRecord
	var raw json.RawMessage

	if err := row.Scan(&record.ID, &raw); err != nil {
		return profileRecord{}, err
	}
	if err := json.Unmarshal(raw, &record.Profile); err != nil {
		return profileRecord{}, err
	}
	return record, nil
}
