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
	userID, err := RequireUserID(c)
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

	_, err = upsertProfileRecord(ctx, userID, payload)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("profile upsert failed: %v", err))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func upsertProfileRecord(ctx context.Context, userID string, profile map[string]interface{}) (models.ProfileRecord, error) {
	db, err := RequireDB()
	if err != nil {
		return models.ProfileRecord{}, err
	}

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return models.ProfileRecord{}, err
	}

const profileUpsertSQL = `
INSERT INTO profiles (user_id, profile, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (user_id) DO UPDATE SET
	profile = excluded.profile,
	updated_at = now()
RETURNING user_id, profile;
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
