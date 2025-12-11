package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/user/Noiseless-service/handlers"
)

func main() {
	loadEnvFile()

	if _, err := initAppDB(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := InitSupabaseJWT(); err != nil {
		log.Fatalf("failed to init Supabase JWT: %v", err)
	}

	handlers.GetDB = getDB
	handlers.UserIDFromCtx = userIDFromCtx
	handlers.HTTPClient = httpClient

	app := fiber.New()
	app.Use(logger.New())
	registerRoutes(app)

	if err := app.Listen(":6469"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
