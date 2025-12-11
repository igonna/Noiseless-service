package handlers

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/user/Noiseless-service/models"
)

type ScrapeFunc func(ctx context.Context, targetURL string) (string, error)
type WebsiteSyncFunc func(ctx context.Context, userID string) (models.ProfileWebsitesResponse, bool, error)
type ProfileUpsertFunc func(ctx context.Context, userID string, profile map[string]interface{}) (models.ProfileRecord, error)
type ProfileFetchFunc func(ctx context.Context, userID string) (models.ProfileRecord, error)
type UserIDProviderFunc func(*fiber.Ctx) (string, error)

type Dependencies struct {
	ScrapeWebsite      ScrapeFunc
	EnsureWebsitesSync WebsiteSyncFunc
	UpsertProfile      ProfileUpsertFunc
	FetchProfile       ProfileFetchFunc
	UserIDProvider     UserIDProviderFunc
	ProfileMissingFunc func(error) bool
}

type Handlers struct {
	deps Dependencies
}

func New(deps Dependencies) (*Handlers, error) {
	switch {
	case deps.ScrapeWebsite == nil:
		return nil, errors.New("ScrapeWebsite dependency is required")
	case deps.EnsureWebsitesSync == nil:
		return nil, errors.New("EnsureWebsitesSync dependency is required")
	case deps.UpsertProfile == nil:
		return nil, errors.New("UpsertProfile dependency is required")
	case deps.FetchProfile == nil:
		return nil, errors.New("FetchProfile dependency is required")
	case deps.UserIDProvider == nil:
		return nil, errors.New("UserIDProvider dependency is required")
	case deps.ProfileMissingFunc == nil:
		return nil, errors.New("ProfileMissingFunc dependency is required")
	}

	return &Handlers{deps: deps}, nil
}
