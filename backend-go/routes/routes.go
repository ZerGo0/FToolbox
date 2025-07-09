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
	creatorHandler := handlers.NewCreatorHandler(db, fanslyClient)
	workerHandler := handlers.NewWorkerHandler(db)

	// Tag routes
	api.Get("/tags", tagHandler.GetTags)
	api.Get("/tags/statistics", tagHandler.GetTagStatistics)
	api.Post("/tags/request", tagHandler.RequestTag)

	// Creator routes
	api.Get("/creators", creatorHandler.GetCreators)
	api.Get("/creators/statistics", creatorHandler.GetCreatorStatistics)
	api.Post("/creators/request", creatorHandler.RequestCreator)

	// Worker routes
	api.Get("/workers/status", workerHandler.GetStatus)

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}
