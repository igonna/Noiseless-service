package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/user/Noiseless-service/api"
	"github.com/user/Noiseless-service/api/handlers"
	"github.com/user/Noiseless-service/internal/db"
	"github.com/user/Noiseless-service/internal/openai"
)

func main() {
	loadEnvFile()

	if _, err := db.Init(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := api.InitSupabaseJWT(); err != nil {
		log.Fatalf("failed to init Supabase JWT: %v", err)
	}

	handlers.GetDB = db.Get
	handlers.UserIDFromCtx = api.UserIDFromCtx
	handlers.HTTPClient = httpClient
	handlers.OpenAIClient = openai.NewClient(httpClient)

	app := fiber.New()
	app.Use(logger.New())
	api.RegisterRoutes(app)

	if err := app.Listen(":6469"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
