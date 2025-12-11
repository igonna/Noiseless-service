package today

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/user/Noiseless-service/api/handlers"
	"github.com/user/Noiseless-service/models"
)

func ensureWebsitesSynced(ctx context.Context, userID string) (models.ProfileWebsitesResponse, bool, error) {
	db, err := handlers.RequireDB()
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

	var resp models.ProfileWebsitesResponse
	if needsRefresh {
		resp, err = fetchProfileWebsites(ctx, string(profile.Profile))
		if err != nil {
			return models.ProfileWebsitesResponse{}, false, err
		}

		buf, err := json.Marshal(resp)
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
	} else {
		resp, err = decodeWebsitesPayload(websitesSnap.Websites)
		if err != nil {
			return models.ProfileWebsitesResponse{}, false, err
		}
	}

	return resp, needsRefresh, nil
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
