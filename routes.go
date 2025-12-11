package main

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func registerRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/signal", handleSignalDefinition)
	v1Signal := app.Group("/v1/signal", SupabaseAuth())
	v1Signal.Get("/today", handleTodaySignals)
	app.Post("/scrape", handleScrape)
	v1User := app.Group("/v1/user", SupabaseAuth())
	v1User.Get("/profile", handleProfileGet)
	v1User.Post("/profile/update", handleProfileUpdate)
}

func handleSignalDefinition(c *fiber.Ctx) error {
	return c.JSON(explanationResponse{
		Topic:      "Signal",
		Definition: "In operating systems, a signal is an asynchronous notification sent to a process or thread to notify it of events such as interrupts, termination requests, or exceptional conditions.",
		KeyConcepts: []string{
			"Signals interrupt the normal flow of execution so a process can react immediately.",
			"Each signal has a numeric identifier and default action defined by the OS.",
			"Processes can register handlers to perform custom logic when a signal arrives.",
		},
	})
}

func handleScrape(c *fiber.Ctx) error {
	var req scrapeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON payload")
	}

	target := strings.TrimSpace(req.URL)
	if target == "" {
		return fiber.NewError(fiber.StatusBadRequest, "url is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	body, err := scrapeWebsite(ctx, target)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	return c.JSON(scrapeResponse{Body: body})
}
