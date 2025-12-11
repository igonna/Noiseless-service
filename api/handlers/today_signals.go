package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/Noiseless-service/models"
)

var errProfileNotFound = errors.New("profile not found")

func TodaySignals(c *fiber.Ctx) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	websitesResp, refreshed, err := ensureWebsitesSynced(ctx, userID)
	if err != nil {
		if errors.Is(err, errProfileNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"signals":            []models.Signal{},
		"websites":           websitesResp.Websites,
		"websites_refreshed": refreshed,
	})
}

func ensureWebsitesSynced(ctx context.Context, userID string) (models.ProfileWebsitesResponse, bool, error) {
	db, err := requireDB()
	if err != nil {
		return models.ProfileWebsitesResponse{}, false, err
	}

	profile, err := fetchProfileSnapshot(ctx, db, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ProfileWebsitesResponse{}, false, errProfileNotFound
		}
		return models.ProfileWebsitesResponse{}, false, err
	}

	websitesSnap, found, err := fetchWebsitesSnapshot(ctx, db, userID)
	if err != nil {
		return models.ProfileWebsitesResponse{}, false, err
	}

	needsRefresh := !found || !websitesSnap.UpdatedAt.Valid || !websitesSnap.UpdatedAt.Time.Equal(profile.UpdatedAt)

	if needsRefresh {
		payload := buildPlaceholderWebsites(profile.Profile)
		buf, err := json.Marshal(payload)
		if err != nil {
			return models.ProfileWebsitesResponse{}, false, err
		}

		if err := saveWebsitesSnapshot(ctx, db, userID, buf, profile.UpdatedAt); err != nil {
			return models.ProfileWebsitesResponse{}, false, err
		}

		websitesSnap = websiteSnapshot{
			UserID:    userID,
			Websites:  buf,
			UpdatedAt: sql.NullTime{Time: profile.UpdatedAt, Valid: true},
		}
	}

	parsed, err := decodeWebsitesPayload(websitesSnap.Websites)
	if err != nil {
		return models.ProfileWebsitesResponse{}, false, err
	}

	return parsed, needsRefresh, nil
}

const (
	profileSnapshotSQL = `
SELECT user_id, profile, updated_at
FROM profiles
WHERE user_id = $1;
`
	websitesSnapshotSQL = `
SELECT user_id, websites, updated_at
FROM websites
WHERE user_id = $1;
`
	websitesUpdateSQL = `
UPDATE websites
SET websites = $1,
    updated_at = $2
WHERE user_id = $3;
`
	websitesInsertSQL = `
INSERT INTO websites (user_id, websites, updated_at)
VALUES ($1, $2, $3);
`
)

func fetchProfileSnapshot(ctx context.Context, db *sql.DB, userID string) (profileSnapshot, error) {
	row := db.QueryRowContext(ctx, profileSnapshotSQL, userID)

	var snap profileSnapshot
	if err := row.Scan(&snap.UserID, &snap.Profile, &snap.UpdatedAt); err != nil {
		return profileSnapshot{}, err
	}

	return snap, nil
}

func fetchWebsitesSnapshot(ctx context.Context, db *sql.DB, userID string) (websiteSnapshot, bool, error) {
	row := db.QueryRowContext(ctx, websitesSnapshotSQL, userID)

	var snap websiteSnapshot
	if err := row.Scan(&snap.UserID, &snap.Websites, &snap.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return websiteSnapshot{}, false, nil
		}
		return websiteSnapshot{}, false, err
	}

	return snap, true, nil
}

func saveWebsitesSnapshot(ctx context.Context, db *sql.DB, userID string, payload []byte, updatedAt time.Time) error {
	res, err := db.ExecContext(ctx, websitesUpdateSQL, payload, updatedAt, userID)
	if err != nil {
		return err
	}

	if rows, err := res.RowsAffected(); err == nil && rows > 0 {
		return nil
	}

	_, err = db.ExecContext(ctx, websitesInsertSQL, userID, payload, updatedAt)
	return err
}

func decodeWebsitesPayload(raw []byte) (models.ProfileWebsitesResponse, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return models.ProfileWebsitesResponse{}, nil
	}

	var resp models.ProfileWebsitesResponse
	if err := json.Unmarshal(trimmed, &resp); err == nil {
		return resp, nil
	}

	var arr []models.WebsiteResource
	if err := json.Unmarshal(trimmed, &arr); err == nil {
		return models.ProfileWebsitesResponse{Websites: arr}, nil
	}

	return models.ProfileWebsitesResponse{}, fmt.Errorf("unable to decode websites payload")
}

func buildPlaceholderWebsites(profile json.RawMessage) models.ProfileWebsitesResponse {
	sum := sha256.Sum256(profile)
	hash := hex.EncodeToString(sum[:])
	domain := fmt.Sprintf("%s.com", hash)

	resource := models.WebsiteResource{
		Name: hash,
		URL:  "https://" + domain,
		Why:  "Auto-generated placeholder derived from profile hash",
	}

	return models.ProfileWebsitesResponse{Websites: []models.WebsiteResource{resource}}
}

type profileSnapshot struct {
	UserID    string
	Profile   json.RawMessage
	UpdatedAt time.Time
}

type websiteSnapshot struct {
	UserID    string
	Websites  json.RawMessage
	UpdatedAt sql.NullTime
}
