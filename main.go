package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	loadEnvFile()

	if err := runMigrations(); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	app := fiber.New()
	registerRoutes(app)

	if err := app.Listen(":6469"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
