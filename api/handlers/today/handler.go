package today

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/Noiseless-service/api/handlers"
	"github.com/user/Noiseless-service/models"
)

var errProfileNotFound = errors.New("profile not found")

// Handle processes GET /v1/signal/today requests.
func Handle(c *fiber.Ctx) error {
	userID, err := handlers.RequireUserID(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	websitesResp, refreshed, err := ensureWebsitesSynced(ctx, userID)
	if err != nil {
		if errors.Is(err, errProfileNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"signals":            []models.Signal{},
		"websites":           websitesResp.Websites,
		"websites_refreshed": refreshed,
	})
}
