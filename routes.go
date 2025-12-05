package main

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func registerRoutes(app *fiber.App) {
	app.Get("/signal", handleSignalDefinition)
	app.Get("/today-signals", handleTodaySignals)
	app.Post("/signal-action", handleSignalAction)
	app.Post("/scrape", handleScrape)
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

func handleTodaySignals(c *fiber.Ctx) error {
	list := listSignalsNewestFirst()
	return c.JSON(fiber.Map{"today-signals": list})
}

func handleSignalAction(c *fiber.Ctx) error {
	var req signalActionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON payload")
	}

	req.SignalID = strings.TrimSpace(req.SignalID)
	if req.SignalID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "signal_id is required")
	}

	snap, found := getSignalCopy(req.SignalID)
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "signal not found")
	}
	if snap.ActionResult != nil {
		return c.JSON(snap)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	result, err := fetchActionResult(ctx, snap)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	updated, err := setSignalActionResult(req.SignalID, result)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(updated)
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
