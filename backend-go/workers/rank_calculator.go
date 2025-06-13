package workers

import (
	"context"
	"ftoolbox/config"
	"ftoolbox/utils"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RankCalculatorWorker struct {
	BaseWorker
	db       *gorm.DB
	interval time.Duration
}

func NewRankCalculatorWorker(db *gorm.DB, cfg *config.Config) *RankCalculatorWorker {
	interval := time.Duration(cfg.RankCalculationInterval) * time.Millisecond

	return &RankCalculatorWorker{
		BaseWorker: NewBaseWorker("rank-calculator", interval),
		db:         db,
		interval:   interval,
	}
}

func (w *RankCalculatorWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Calculate ranks immediately on startup
	w.calculateRanks()

	for {
		select {
		case <-ctx.Done():
			zap.L().Info("Rank calculator worker stopping")
			return ctx.Err()
		case <-ticker.C:
			w.calculateRanks()
		}
	}
}

func (w *RankCalculatorWorker) calculateRanks() {
	startTime := time.Now()
	zap.L().Info("Starting rank calculation")

	if err := utils.CalculateTagRanks(w.db); err != nil {
		zap.L().Error("Failed to calculate ranks", zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	zap.L().Info("Rank calculation completed",
		zap.Duration("duration", duration))
}
