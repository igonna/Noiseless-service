package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/user/Noiseless-service/internal/openai"
)

var (
	// GetDB should be assigned by main so handlers can reuse the shared *sql.DB pool.
	GetDB func() (*sql.DB, error)
	// UserIDFromCtx extracts the authenticated user ID from the Fiber context.
	UserIDFromCtx func(*fiber.Ctx) (string, error)
	// HTTPClient is used by handlers that need outbound HTTP calls (e.g., scraping).
	HTTPClient *http.Client = http.DefaultClient
	// OpenAIClient provides access to the shared OpenAI caller.
	OpenAIClient *openai.Client
)

func RequireDB() (*sql.DB, error) {
	if GetDB == nil {
		return nil, errors.New("handlers: GetDB is not configured")
	}
	return GetDB()
}

func RequireUserID(c *fiber.Ctx) (string, error) {
	if UserIDFromCtx == nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, "handlers: user context helper not configured")
	}
	return UserIDFromCtx(c)
}

func RequireOpenAIClient() (*openai.Client, error) {
	if OpenAIClient == nil {
		return nil, errors.New("handlers: OpenAI client not configured")
	}
	return OpenAIClient, nil
}
