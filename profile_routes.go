package main

import (
	"context"
	"encoding/json"

	"github.com/user/Noiseless-service/models"
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

func upsertProfileRecord(ctx context.Context, userID string, profile map[string]interface{}) (models.ProfileRecord, error) {
	db, err := getDB()
	if err != nil {
		return models.ProfileRecord{}, err
	}

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return models.ProfileRecord{}, err
	}

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

func fetchProfileRecord(ctx context.Context, userID string) (models.ProfileRecord, error) {
	db, err := getDB()
	if err != nil {
		return models.ProfileRecord{}, err
	}

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
