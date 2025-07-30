package main

import (
	"ftoolbox/config"
	"ftoolbox/database"
	"ftoolbox/fansly"
	"ftoolbox/models"
	"ftoolbox/routes"
	"ftoolbox/utils"
	"ftoolbox/workers"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	// Initialize zap logger
	var logger *zap.Logger
	var err error

	// Always use development config for better formatting
	zapConfig := zap.NewDevelopmentConfig()

	// Set the log level based on LOG_LEVEL env var
	switch cfg.LogLevel {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err = zapConfig.Build()
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
		if err := utils.CalculateTagRanks(db); err != nil {
			zap.L().Error("Failed to calculate initial tag ranks", zap.Error(err))
		}
	}

	// Calculate initial heat scores
	var heatCount int64
	db.Model(&models.Tag{}).Where("heat = 0").Count(&heatCount)
	if heatCount > 0 {
		zap.L().Info("Calculating initial heat scores for tags")
		if err := utils.CalculateTagHeatScores(db); err != nil {
			zap.L().Error("Failed to calculate initial tag heat scores", zap.Error(err))
		}
	}

	// Calculate initial ranks for creators if needed
	var creatorCount int64
	db.Model(&models.Creator{}).Where("rank IS NULL").Count(&creatorCount)
	if creatorCount > 0 {
		zap.L().Info("Calculating initial ranks for creators")
		if err := utils.CalculateCreatorRanks(db); err != nil {
			zap.L().Error("Failed to calculate initial creator ranks", zap.Error(err))
		}
	}

	// Initialize Fansly client with global rate limiting
	fanslyClient := fansly.NewClient()

	// Configure global rate limit
	fanslyClient.SetGlobalRateLimit(cfg.GlobalRateLimit, cfg.GlobalRateLimitWindow)
	zap.L().Info("Configured global rate limit",
		zap.Int("max_requests", cfg.GlobalRateLimit),
		zap.Int("window_seconds", cfg.GlobalRateLimitWindow))

	// Initialize worker manager
	workerManager := workers.NewWorkerManager(db, cfg.WorkerEnabled)

	// Register workers
	tagUpdater := workers.NewTagUpdaterWorker(db, cfg, fanslyClient)
	tagDiscovery := workers.NewTagDiscoveryWorker(db, cfg, fanslyClient)
	rankCalculator := workers.NewRankCalculatorWorker(db, cfg)
	creatorUpdater := workers.NewCreatorUpdaterWorker(db, fanslyClient)
	statisticsCalculator := workers.NewStatisticsCalculatorWorker(db, cfg)

	if err := workerManager.Register(tagUpdater); err != nil {
		zap.L().Error("Failed to register tag updater", zap.Error(err))
	}
	if err := workerManager.Register(tagDiscovery); err != nil {
		zap.L().Error("Failed to register tag discovery", zap.Error(err))
	}
	if err := workerManager.Register(rankCalculator); err != nil {
		zap.L().Error("Failed to register rank calculator", zap.Error(err))
	}
	if err := workerManager.Register(creatorUpdater); err != nil {
		zap.L().Error("Failed to register creator updater", zap.Error(err))
	}
	if err := workerManager.Register(statisticsCalculator); err != nil {
		zap.L().Error("Failed to register statistics calculator", zap.Error(err))
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
			if err := workerManager.Start("rank-calculator"); err != nil {
				zap.L().Error("Failed to start rank calculator", zap.Error(err))
			}
			if err := workerManager.Start("creator-updater"); err != nil {
				zap.L().Error("Failed to start creator updater", zap.Error(err))
			}
			if err := workerManager.Start("statistics-calculator"); err != nil {
				zap.L().Error("Failed to start statistics calculator", zap.Error(err))
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
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(etag.New())

	routes.Setup(app, db, workerManager, fanslyClient)

	zap.L().Info("Server starting", zap.String("port", cfg.Port))
	if err := app.Listen(":" + cfg.Port); err != nil {
		zap.L().Fatal("Failed to start server", zap.Error(err))
	}
}
