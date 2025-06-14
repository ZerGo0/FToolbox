package routes

import (
	"ftoolbox/fansly"
	"ftoolbox/handlers"
	"ftoolbox/workers"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB, workerManager *workers.WorkerManager, fanslyClient *fansly.Client) {
	api := app.Group("/api")

	tagHandler := handlers.NewTagHandler(db, fanslyClient)
	workerHandler := handlers.NewWorkerHandler(db)

	// Tag routes
	api.Get("/tags", tagHandler.GetTags)
	api.Post("/tags/request", tagHandler.RequestTag)

	// Worker routes
	api.Get("/workers/status", workerHandler.GetStatus)

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Rate limit stats (for monitoring)
	api.Get("/ratelimits/stats", func(c *fiber.Ctx) error {
		stats := fanslyClient.GetRateLimitStats()
		return c.JSON(fiber.Map{
			"endpoints": stats,
			"adaptive":  true,
		})
	})
}
