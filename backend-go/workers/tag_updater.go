package workers

import (
	"context"
	"fmt"
	"ftoolbox/config"
	"ftoolbox/fansly"
	"ftoolbox/models"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TagUpdaterWorker struct {
	BaseWorker
	db     *gorm.DB
	client *fansly.Client
}

func NewTagUpdaterWorker(db *gorm.DB, cfg *config.Config) *TagUpdaterWorker {
	interval := time.Duration(cfg.WorkerUpdateInterval) * time.Millisecond

	return &TagUpdaterWorker{
		BaseWorker: NewBaseWorker("tag-updater", interval),
		db:         db,
		client:     fansly.NewClient(cfg.FanslyAPIRateLimit),
	}
}

func (w *TagUpdaterWorker) Run(ctx context.Context) error {
	zap.L().Info("Running tag updater")

	// Get tags that need updating (haven't been checked in 24 hours)
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	var tags []models.Tag
	if err := w.db.Where("last_checked_at IS NULL OR last_checked_at < ?", twentyFourHoursAgo).
		Order("view_count DESC").
		Limit(20).
		Find(&tags).Error; err != nil {
		return fmt.Errorf("failed to fetch tags: %w", err)
	}

	if len(tags) == 0 {
		zap.L().Debug("No tags need updating")
		return nil
	}

	zap.L().Info("Updating tags", zap.Int("count", len(tags)))

	for _, tag := range tags {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := w.updateTag(&tag); err != nil {
			zap.L().Error("Failed to update tag",
				zap.String("tag", tag.Tag),
				zap.Error(err))

			// Update last_checked_at even on error to avoid immediate retries
			now := time.Now()
			tag.LastCheckedAt = &now
			if updateErr := w.db.Save(&tag).Error; updateErr != nil {
				zap.L().Error("Failed to update last checked time after error",
					zap.String("tag", tag.Tag),
					zap.Error(updateErr))
			}
			continue
		}
	}

	return nil
}

func (w *TagUpdaterWorker) updateTag(tag *models.Tag) error {
	// Fetch current view count from Fansly
	viewCount, err := w.client.FetchTagViewCount(tag.Tag)
	if err != nil {
		return fmt.Errorf("failed to fetch view count: %w", err)
	}

	// Calculate change
	change := viewCount - tag.ViewCount

	// Start transaction
	tx := w.db.Begin()

	// Update tag
	now := time.Now()
	tag.ViewCount = viewCount
	tag.LastCheckedAt = &now
	tag.UpdatedAt = now

	if err := tx.Save(tag).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update tag: %w", err)
	}

	history := models.TagHistory{
		TagID:     tag.ID,
		ViewCount: viewCount,
		Change:    change,
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create history: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	zap.L().Debug("Updated tag",
		zap.String("tag", tag.Tag),
		zap.Int64("viewCount", viewCount),
		zap.Int64("change", change))

	return nil
}
