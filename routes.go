package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/user/Noiseless-service/handlers"
)

func registerRoutes(app *fiber.App, h *handlers.Handlers) {
	if h == nil {
		panic("handlers must not be nil")
	}

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/signal", h.SignalDefinition)
	v1Signal := app.Group("/v1/signal", SupabaseAuth())
	v1Signal.Get("/today", h.TodaySignals)
	app.Post("/scrape", h.Scrape)
	v1User := app.Group("/v1/user", SupabaseAuth())
	v1User.Get("/profile", h.ProfileGet)
	v1User.Post("/profile/update", h.ProfileUpdate)
}
