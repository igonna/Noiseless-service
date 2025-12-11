package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/user/Noiseless-service/api/handlers"
)

func RegisterRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/signal", handlers.SignalDefinition)
	v1Signal := app.Group("/v1/signal", SupabaseAuth())
	v1Signal.Get("/today", handlers.TodaySignals)
	app.Post("/scrape", handlers.Scrape)
	v1User := app.Group("/v1/user", SupabaseAuth())
	v1User.Get("/profile", handlers.ProfileGet)
	v1User.Post("/profile/update", handlers.ProfileUpdate)
}
