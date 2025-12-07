package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
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

	app := fiber.New()
	app.Use(logger.New())
	registerRoutes(app)

	if err := app.Listen(":6469"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
