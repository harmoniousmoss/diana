package routes

import (
	"myfiberproject/handlers"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures API endpoints.
func SetupRoutes(app *fiber.App) {
	// Get email
	app.Get("/store-media-monitoring", handlers.StoreMediaMonitoring)
}
