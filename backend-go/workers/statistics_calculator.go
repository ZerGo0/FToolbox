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

type StatisticsCalculatorWorker struct {
	BaseWorker
	db *gorm.DB
}

func NewStatisticsCalculatorWorker(db *gorm.DB, cfg *config.Config) *StatisticsCalculatorWorker {
	// Default to 1 hour interval
	interval := time.Hour
	if cfg.WorkerStatisticsInterval > 0 {
		interval = time.Duration(cfg.WorkerStatisticsInterval) * time.Millisecond
	}

	return &StatisticsCalculatorWorker{
		BaseWorker: NewBaseWorker("statistics-calculator", interval),
		db:         db,
	}
}

func (w *StatisticsCalculatorWorker) Run(ctx context.Context) error {
	zap.L().Info("Running statistics calculator")

	// Calculate tag statistics
	if err := w.calculateTagStatistics(ctx); err != nil {
		zap.L().Error("Failed to calculate tag statistics", zap.Error(err))
		// Continue with creator statistics even if tag statistics failed
	}

	// Calculate creator statistics
	if err := w.calculateCreatorStatistics(ctx); err != nil {
		zap.L().Error("Failed to calculate creator statistics", zap.Error(err))
		return err
	}

	return nil
}

func (w *StatisticsCalculatorWorker) calculateTagStatistics(ctx context.Context) error {
	// Start a transaction
	tx := w.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Calculate current total view count and post count
	var totalViewCount, totalPostCount int64
	if err := tx.Model(&models.Tag{}).
		Where("is_deleted = ?", false).
		Select("COALESCE(SUM(view_count), 0), COALESCE(SUM(post_count), 0)").
		Row().Scan(&totalViewCount, &totalPostCount); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to calculate total counts: %w", err)
	}

	// Get the previous statistics record from 24 hours ago
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	var previousStats models.TagStatistics
	if err := tx.Where("calculated_at <= ?", twentyFourHoursAgo).
		Order("calculated_at DESC").
		First(&previousStats).Error; err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("failed to fetch previous statistics: %w", err)
	}

	// Calculate 24-hour changes based on view count
	var change24h int64
	var changePercent24h float64
	var postChange24h int64
	var postChangePercent24h float64
	if previousStats.ID != 0 {
		change24h = totalViewCount - previousStats.TotalViewCount
		if previousStats.TotalViewCount > 0 {
			changePercent24h = (float64(change24h) / float64(previousStats.TotalViewCount)) * 100
		}
		postChange24h = totalPostCount - previousStats.TotalPostCount
		if previousStats.TotalPostCount > 0 {
			postChangePercent24h = (float64(postChange24h) / float64(previousStats.TotalPostCount)) * 100
		}
	}

	// Create new statistics record
	newStats := models.TagStatistics{
		TotalViewCount:       totalViewCount,
		TotalPostCount:       totalPostCount,
		Change24h:            change24h,
		ChangePercent24h:     changePercent24h,
		PostChange24h:        postChange24h,
		PostChangePercent24h: postChangePercent24h,
		CalculatedAt:         time.Now(),
	}

	if err := tx.Create(&newStats).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create tag statistics: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit tag statistics: %w", err)
	}

	zap.L().Info("Tag statistics calculated successfully",
		zap.Int64("totalViewCount", totalViewCount),
		zap.Int64("totalPostCount", totalPostCount),
		zap.Int64("change24h", change24h),
		zap.Float64("changePercent24h", changePercent24h),
		zap.Int64("postChange24h", postChange24h),
		zap.Float64("postChangePercent24h", postChangePercent24h))

	return nil
}

func (w *StatisticsCalculatorWorker) calculateCreatorStatistics(ctx context.Context) error {
	// Start a transaction
	tx := w.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Calculate current totals
	var totalFollowers, totalMediaLikes, totalPostLikes int64
	if err := tx.Model(&models.Creator{}).
		Where("is_deleted = ?", false).
		Select("COALESCE(SUM(followers), 0), COALESCE(SUM(media_likes), 0), COALESCE(SUM(post_likes), 0)").
		Row().Scan(&totalFollowers, &totalMediaLikes, &totalPostLikes); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to calculate creator totals: %w", err)
	}

	// Get the previous statistics record from 24 hours ago
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	var previousStats models.CreatorStatistics
	if err := tx.Where("calculated_at <= ?", twentyFourHoursAgo).
		Order("calculated_at DESC").
		First(&previousStats).Error; err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("failed to fetch previous creator statistics: %w", err)
	}

	// Calculate 24-hour changes
	var followersChange24h, mediaLikesChange24h, postLikesChange24h int64
	var followersChangePercent24h, mediaLikesChangePercent24h, postLikesChangePercent24h float64

	if previousStats.ID != 0 {
		// Followers
		followersChange24h = totalFollowers - previousStats.TotalFollowers
		if previousStats.TotalFollowers > 0 {
			followersChangePercent24h = (float64(followersChange24h) / float64(previousStats.TotalFollowers)) * 100
		}

		// Media Likes
		mediaLikesChange24h = totalMediaLikes - previousStats.TotalMediaLikes
		if previousStats.TotalMediaLikes > 0 {
			mediaLikesChangePercent24h = (float64(mediaLikesChange24h) / float64(previousStats.TotalMediaLikes)) * 100
		}

		// Post Likes
		postLikesChange24h = totalPostLikes - previousStats.TotalPostLikes
		if previousStats.TotalPostLikes > 0 {
			postLikesChangePercent24h = (float64(postLikesChange24h) / float64(previousStats.TotalPostLikes)) * 100
		}
	}

	// Create new statistics record
	newStats := models.CreatorStatistics{
		TotalFollowers:             totalFollowers,
		FollowersChange24h:         followersChange24h,
		FollowersChangePercent24h:  followersChangePercent24h,
		TotalMediaLikes:            totalMediaLikes,
		MediaLikesChange24h:        mediaLikesChange24h,
		MediaLikesChangePercent24h: mediaLikesChangePercent24h,
		TotalPostLikes:             totalPostLikes,
		PostLikesChange24h:         postLikesChange24h,
		PostLikesChangePercent24h:  postLikesChangePercent24h,
		CalculatedAt:               time.Now(),
	}

	if err := tx.Create(&newStats).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create creator statistics: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit creator statistics: %w", err)
	}

	zap.L().Info("Creator statistics calculated successfully",
		zap.Int64("totalFollowers", totalFollowers),
		zap.Int64("totalMediaLikes", totalMediaLikes),
		zap.Int64("totalPostLikes", totalPostLikes))

	return nil
}
