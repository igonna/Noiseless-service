package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	registerRoutes(app)

	if err := app.Listen(":6469"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
