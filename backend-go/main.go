package main

import (
	"ftoolbox/config"
	"ftoolbox/database"
	"ftoolbox/fansly"
	"ftoolbox/models"
	"ftoolbox/routes"
	"ftoolbox/workers"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	// Initialize zap logger
	var logger *zap.Logger
	var err error
	if cfg.LogLevel == "debug" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	db, err := database.Connect(cfg)
	if err != nil {
		zap.L().Fatal("Failed to connect to database", zap.Error(err))
	}

	if err := database.AutoMigrate(db); err != nil {
		zap.L().Fatal("Failed to run migrations", zap.Error(err))
	}

	// Calculate initial ranks if needed
	var tagCount int64
	db.Model(&models.Tag{}).Where("rank IS NULL").Count(&tagCount)
	if tagCount > 0 {
		zap.L().Info("Calculating initial ranks for tags")
		if err := models.CalculateTagRanks(db); err != nil {
			zap.L().Error("Failed to calculate initial ranks", zap.Error(err))
		}
	}

	// Initialize Fansly client
	fanslyClient := fansly.NewClient(cfg.FanslyAPIRateLimit)

	// Initialize worker manager
	workerManager := workers.NewWorkerManager(db, cfg.WorkerEnabled)

	// Register workers
	tagUpdater := workers.NewTagUpdaterWorker(db, cfg)
	tagDiscovery := workers.NewTagDiscoveryWorker(db, cfg)

	if err := workerManager.Register(tagUpdater); err != nil {
		zap.L().Error("Failed to register tag updater", zap.Error(err))
	}
	if err := workerManager.Register(tagDiscovery); err != nil {
		zap.L().Error("Failed to register tag discovery", zap.Error(err))
	}

	// Start workers if enabled
	if cfg.WorkerEnabled {
		go func() {
			if err := workerManager.Start("tag-updater"); err != nil {
				zap.L().Error("Failed to start tag updater", zap.Error(err))
			}
			if err := workerManager.Start("tag-discovery"); err != nil {
				zap.L().Error("Failed to start tag discovery", zap.Error(err))
			}
		}()
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		zap.L().Info("Received shutdown signal, stopping workers...")
		workerManager.StopAll()
		os.Exit(0)
	}()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, HEAD, PUT, PATCH, POST, DELETE",
	}))

	routes.Setup(app, db, workerManager, fanslyClient)

	zap.L().Info("Server starting", zap.String("port", cfg.Port))
	if err := app.Listen(":" + cfg.Port); err != nil {
		zap.L().Fatal("Failed to start server", zap.Error(err))
	}
}
