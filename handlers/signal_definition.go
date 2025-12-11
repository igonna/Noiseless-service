package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/user/Noiseless-service/models"
)

func (h *Handlers) SignalDefinition(c *fiber.Ctx) error {
	return c.JSON(models.ExplanationResponse{
		Topic:      "Signal",
		Definition: "In operating systems, a signal is an asynchronous notification sent to a process or thread to notify it of events such as interrupts, termination requests, or exceptional conditions.",
		KeyConcepts: []string{
			"Signals interrupt the normal flow of execution so a process can react immediately.",
			"Each signal has a numeric identifier and default action defined by the OS.",
			"Processes can register handlers to perform custom logic when a signal arrives.",
		},
	})
}
