package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/Noiseless-service/models"
)

func ProfileGet(c *fiber.Ctx) error {
	userID, err := RequireUserID(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	record, err := fetchProfileRecord(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "profile not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"profile": record})
}

func fetchProfileRecord(ctx context.Context, userID string) (models.ProfileRecord, error) {
	db, err := RequireDB()
	if err != nil {
		return models.ProfileRecord{}, err
	}

	const profileSelectSQL = `
SELECT id, profile
FROM profiles
WHERE id = $1;
`

	row := db.QueryRowContext(ctx, profileSelectSQL, userID)
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
