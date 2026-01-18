package workers

import (
	"context"
	"fmt"
	"ftoolbox/config"
	"ftoolbox/models"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TagCleanupWorker struct {
	BaseWorker
	db       *gorm.DB
	minViews int64
}

func NewTagCleanupWorker(db *gorm.DB, cfg *config.Config) *TagCleanupWorker {
	interval := time.Duration(cfg.WorkerTagCleanupInterval) * time.Millisecond

	return &TagCleanupWorker{
		BaseWorker: NewBaseWorker("tag-cleanup", interval),
		db:         db,
		minViews:   500,
	}
}

func (w *TagCleanupWorker) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	zap.L().Info("Running tag cleanup", zap.Int64("minViews", w.minViews))

	var tagCount int64
	if err := w.db.Model(&models.Tag{}).
		Where("view_count < ?", w.minViews).
		Count(&tagCount).Error; err != nil {
		return fmt.Errorf("failed to count tags for cleanup: %w", err)
	}

	if tagCount == 0 {
		zap.L().Debug("No tags to cleanup")
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tx := w.db.Begin()
	tagSubquery := tx.Model(&models.Tag{}).
		Select("id").
		Where("view_count < ?", w.minViews)

	if err := tx.Where("tag_id IN (?)", tagSubquery).Delete(&models.TagHistory{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete tag history: %w", err)
	}

	if err := tx.Where("tag_id IN (?) OR related_tag_id IN (?)", tagSubquery, tagSubquery).
		Delete(&models.TagRelationDaily{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete tag relations: %w", err)
	}

	result := tx.Where("view_count < ?", w.minViews).Delete(&models.Tag{})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete tags: %w", result.Error)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit tag cleanup: %w", err)
	}

	zap.L().Info("Tag cleanup completed", zap.Int64("deleted", result.RowsAffected))

	return nil
}
