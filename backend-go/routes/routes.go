package routes

import (
	"ftoolbox/fansly"
	"ftoolbox/handlers"
	"ftoolbox/workers"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB, workerManager *workers.WorkerManager, fanslyClient *fansly.Client) {
	api := app.Group("/api")

	tagHandler := handlers.NewTagHandler(db, fanslyClient)
	creatorHandler := handlers.NewCreatorHandler(db, fanslyClient)
	workerHandler := handlers.NewWorkerHandler(db)

	// Tag routes
	api.Get("/tags", tagHandler.GetTags)
	api.Get("/tags/banned", tagHandler.GetBannedTags)
	api.Get("/tags/statistics", tagHandler.GetTagStatistics)
	api.Get("/tags/related", tagHandler.GetRelatedTags)
	api.Use("/tags/request", limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
	}))
	api.Post("/tags/request", tagHandler.RequestTag)

	// Creator routes
	api.Get("/creators", creatorHandler.GetCreators)
	api.Get("/creators/statistics", creatorHandler.GetCreatorStatistics)
	api.Use("/creators/request", limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
	}))
	api.Post("/creators/request", creatorHandler.RequestCreator)

	// Worker routes
	api.Get("/workers/status", workerHandler.GetStatus)

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}
