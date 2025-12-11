package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/Noiseless-service/models"
)

func ProfileUpdate(c *fiber.Ctx) error {
	userID, err := requireUserID(c)
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

	record, err := upsertProfileRecord(ctx, userID, payload)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("profile upsert failed: %v", err))
	}

	return c.JSON(fiber.Map{"profile": record})
}

func upsertProfileRecord(ctx context.Context, userID string, profile map[string]interface{}) (models.ProfileRecord, error) {
	db, err := requireDB()
	if err != nil {
		return models.ProfileRecord{}, err
	}

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return models.ProfileRecord{}, err
	}

	const profileUpsertSQL = `
INSERT INTO profiles (id, profile)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET
	profile = excluded.profile
RETURNING id, profile;
`

	row := db.QueryRowContext(ctx, profileUpsertSQL, userID, profileJSON)

	var record models.ProfileRecord
	var raw json.RawMessage

	if err := row.Scan(&record.ID, &raw); err != nil {
		return models.ProfileRecord{}, err
	}

	if err := json.Unmarshal(raw, &record.Profile); err != nil {
		return models.ProfileRecord{}, err
	}

	return record, nil
}
