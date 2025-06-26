package workers

import (
	"context"
	"errors"
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

func NewTagUpdaterWorker(db *gorm.DB, cfg *config.Config, client *fansly.Client) *TagUpdaterWorker {
	interval := time.Duration(cfg.WorkerUpdateInterval) * time.Millisecond

	return &TagUpdaterWorker{
		BaseWorker: NewBaseWorker("tag-updater", interval),
		db:         db,
		client:     client,
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

		if err := w.updateTag(ctx, &tag); err != nil {
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

func (w *TagUpdaterWorker) updateTag(ctx context.Context, tag *models.Tag) error {
	// Fetch current view count from Fansly
	viewCount, err := w.client.GetTagWithContext(ctx, tag.Tag)
	if err != nil {
		// Check if tag no longer exists on Fansly
		if errors.Is(err, fansly.ErrTagNotFound) {
			// Mark tag as deleted
			now := time.Now()
			tag.LastCheckedAt = &now
			tag.UpdatedAt = now

			// Only update deletion fields if not already marked as deleted
			if !tag.IsDeleted {
				tag.IsDeleted = true
				tag.DeletedDetectedAt = &now

				zap.L().Info("Tag no longer exists on Fansly, marking as deleted",
					zap.String("tag", tag.Tag))
			}

			// Save the updated tag (no history entry for deleted tags)
			if err := w.db.Save(tag).Error; err != nil {
				return fmt.Errorf("failed to update deleted tag: %w", err)
			}

			return nil
		}
		return fmt.Errorf("failed to fetch view count: %w", err)
	}

	// If tag was previously deleted but now exists again, clear the deletion flag
	if tag.IsDeleted {
		tag.IsDeleted = false
		tag.DeletedDetectedAt = nil
		zap.L().Info("Tag exists again on Fansly, clearing deleted status",
			zap.String("tag", tag.Tag))
	}

	// Calculate changes
	viewCountChange := viewCount.MediaOfferSuggestionTag.ViewCount - tag.ViewCount

	// Start transaction
	tx := w.db.Begin()

	// Update tag
	now := time.Now()
	tag.ViewCount = viewCount.MediaOfferSuggestionTag.ViewCount
	tag.LastCheckedAt = &now
	tag.UpdatedAt = now

	if err := tx.Save(tag).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update tag: %w", err)
	}

	history := models.TagHistory{
		TagID:     tag.ID,
		ViewCount: viewCount.MediaOfferSuggestionTag.ViewCount,
		Change:    viewCountChange,
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
		zap.Int64("viewCount", viewCount.MediaOfferSuggestionTag.ViewCount),
		zap.Int64("viewCountChange", viewCountChange))

	return nil
}
