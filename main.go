package main

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/user/Noiseless-service/handlers"
)

func main() {
	loadEnvFile()

	if err := runMigrations(); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	if _, err := initAppDB(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := InitSupabaseJWT(); err != nil {
		log.Fatalf("failed to init Supabase JWT: %v", err)
	}

	handlerSet, err := handlers.New(handlers.Dependencies{
		ScrapeWebsite:      scrapeWebsite,
		EnsureWebsitesSync: ensureWebsitesSynced,
		UpsertProfile:      upsertProfileRecord,
		FetchProfile:       fetchProfileRecord,
		UserIDProvider:     userIDFromCtx,
		ProfileMissingFunc: func(err error) bool {
			return errors.Is(err, errProfileNotFound)
		},
	})
	if err != nil {
		log.Fatalf("failed to init handlers: %v", err)
	}

	app := fiber.New()
	app.Use(logger.New())
	registerRoutes(app, handlerSet)

	if err := app.Listen(":6469"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
