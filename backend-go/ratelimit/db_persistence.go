package ratelimit

import (
	"ftoolbox/models"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DBPersistence provides database-based persistence for rate limit configurations
type DBPersistence struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewDBPersistence creates a new database persistence handler
func NewDBPersistence(db *gorm.DB, logger *zap.Logger) *DBPersistence {
	return &DBPersistence{
		db:     db,
		logger: logger,
	}
}

// Save writes the endpoint configurations to the database
func (p *DBPersistence) Save(endpoints map[string]*EndpointConfig) error {
	// Use a transaction for atomic updates
	return p.db.Transaction(func(tx *gorm.DB) error {
		for endpoint, config := range endpoints {
			rateLimit := models.RateLimit{
				ID:                endpoint,
				Limit:             config.Limit,
				Window:            int64(config.Window.Seconds()),
				SuccessStreak:     config.SuccessStreak,
				RateLimitHits:     config.RateLimitHits,
				BackoffMultiplier: config.BackoffMultiplier,
				LastRateLimitHit:  config.LastRateLimitHit,
			}

			// Use Save to create or update
			if err := tx.Save(&rateLimit).Error; err != nil {
				p.logger.Error("Failed to save rate limit config",
					zap.String("endpoint", endpoint),
					zap.Error(err))
				return err
			}
		}
		return nil
	})
}

// Load reads the endpoint configurations from the database
func (p *DBPersistence) Load() (map[string]*EndpointConfig, error) {
	var rateLimits []models.RateLimit
	if err := p.db.Find(&rateLimits).Error; err != nil {
		p.logger.Error("Failed to load rate limit configs", zap.Error(err))
		return nil, err
	}

	endpoints := make(map[string]*EndpointConfig)
	for _, rl := range rateLimits {
		endpoints[rl.ID] = &EndpointConfig{
			Limit:             rl.Limit,
			Window:            time.Duration(rl.Window) * time.Second,
			SuccessStreak:     rl.SuccessStreak,
			RateLimitHits:     rl.RateLimitHits,
			BackoffMultiplier: rl.BackoffMultiplier,
			LastRateLimitHit:  rl.LastRateLimitHit,
			RequestTimestamps: make([]time.Time, 0),
			// Don't restore current backoff - let it start fresh
		}
	}

	p.logger.Info("Loaded rate limit configurations from database",
		zap.Int("count", len(endpoints)))

	return endpoints, nil
}
