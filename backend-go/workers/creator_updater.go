package workers

import (
	"context"
	"fmt"
	"ftoolbox/fansly"
	"ftoolbox/models"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CreatorUpdaterWorker struct {
	BaseWorker
	db     *gorm.DB
	client *fansly.Client
}

func NewCreatorUpdaterWorker(db *gorm.DB, client *fansly.Client) *CreatorUpdaterWorker {
	return &CreatorUpdaterWorker{
		BaseWorker: NewBaseWorker("creator-updater", 10*time.Second),
		db:         db,
		client:     client,
	}
}

func (w *CreatorUpdaterWorker) Run(ctx context.Context) error {
	zap.L().Info("Running creator updater")

	// Get creators that need updating (haven't been checked in 24 hours)
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	var creatorIDs []string
	if err := w.db.Model(&models.Creator{}).
		Where("last_checked_at IS NULL OR last_checked_at < ?", twentyFourHoursAgo).
		Order("followers DESC").
		Limit(25).
		Pluck("id", &creatorIDs).Error; err != nil {
		return fmt.Errorf("failed to fetch creator IDs: %w", err)
	}

	if len(creatorIDs) == 0 {
		zap.L().Debug("No creators need updating")
		return nil
	}

	zap.L().Info("Updating creators", zap.Int("count", len(creatorIDs)))

	// Fetch account data from Fansly API
	accounts, err := w.client.GetAccountsWithContext(ctx, creatorIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch creator accounts: %w", err)
	}

	// Process the fetched accounts
	return w.ProcessCreators(accounts)
}

func (w *CreatorUpdaterWorker) ProcessCreators(accounts []fansly.FanslyAccount) error {
	if len(accounts) == 0 {
		zap.L().Debug("No creators to process")
		return nil
	}

	zap.L().Info("Processing creators", zap.Int("accounts", len(accounts)))

	newCreators := 0
	updatedCreators := 0
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	for _, account := range accounts {
		// Check if creator already exists
		var existingCreator models.Creator
		err := w.db.Where("id = ?", account.ID).First(&existingCreator).Error

		if err == nil {
			// Creator exists - check if it needs updating
			if existingCreator.LastCheckedAt != nil && existingCreator.LastCheckedAt.After(twentyFourHoursAgo) {
				// Already updated within 24 hours, skip
				continue
			}

			// Update existing creator
			if err := w.updateCreator(&existingCreator, &account); err != nil {
				zap.L().Error("Failed to update creator",
					zap.String("username", account.Username),
					zap.Error(err))
				continue
			}
			updatedCreators++
		} else if err == gorm.ErrRecordNotFound {
			// Create new creator
			if err := w.createCreator(&account); err != nil {
				zap.L().Error("Failed to create creator",
					zap.String("username", account.Username),
					zap.Error(err))
				continue
			}
			newCreators++
		} else {
			// Database error
			zap.L().Error("Failed to query creator",
				zap.String("account_id", account.ID),
				zap.Error(err))
			continue
		}
	}

	zap.L().Info("Creator processing completed",
		zap.Int("processed", len(accounts)),
		zap.Int("new", newCreators),
		zap.Int("updated", updatedCreators))

	return nil
}

func (w *CreatorUpdaterWorker) createCreator(account *fansly.FanslyAccount) error {
	// Start transaction
	tx := w.db.Begin()

	// Create new creator
	displayName := account.DisplayName
	if displayName == "" {
		displayName = account.Username
	}

	now := time.Now()
	newCreator := models.Creator{
		ID:            account.ID,
		Username:      account.Username,
		DisplayName:   &displayName,
		MediaLikes:    account.AccountMediaLikes,
		PostLikes:     account.PostLikes,
		Followers:     account.FollowCount,
		ImageCount:    account.TimelineStats.ImageCount,
		VideoCount:    account.TimelineStats.VideoCount,
		LastCheckedAt: &now,
	}

	if err := tx.Create(&newCreator).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create initial history entry
	history := models.CreatorHistory{
		CreatorID:  account.ID,
		MediaLikes: account.AccountMediaLikes,
		PostLikes:  account.PostLikes,
		Followers:  account.FollowCount,
		ImageCount: account.TimelineStats.ImageCount,
		VideoCount: account.TimelineStats.VideoCount,
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	zap.L().Info("Created new creator",
		zap.String("username", account.Username),
		zap.Int64("followers", account.FollowCount))

	return nil
}

func (w *CreatorUpdaterWorker) updateCreator(creator *models.Creator, account *fansly.FanslyAccount) error {
	// If creator was previously deleted but now exists again, clear the deletion flag
	if creator.IsDeleted {
		creator.IsDeleted = false
		creator.DeletedDetectedAt = nil
		zap.L().Info("Creator exists again, clearing deleted status",
			zap.String("username", creator.Username))
	}

	// Start transaction
	tx := w.db.Begin()

	// Update creator
	now := time.Now()
	displayName := account.DisplayName
	if displayName == "" {
		displayName = account.Username
	}

	creator.Username = account.Username
	creator.DisplayName = &displayName
	creator.MediaLikes = account.AccountMediaLikes
	creator.PostLikes = account.PostLikes
	creator.Followers = account.FollowCount
	creator.ImageCount = account.TimelineStats.ImageCount
	creator.VideoCount = account.TimelineStats.VideoCount
	creator.LastCheckedAt = &now
	creator.UpdatedAt = now

	if err := tx.Save(creator).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create history entry
	history := models.CreatorHistory{
		CreatorID:  creator.ID,
		MediaLikes: account.AccountMediaLikes,
		PostLikes:  account.PostLikes,
		Followers:  account.FollowCount,
		ImageCount: account.TimelineStats.ImageCount,
		VideoCount: account.TimelineStats.VideoCount,
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	zap.L().Debug("Updated creator",
		zap.String("username", creator.Username),
		zap.Int64("followers", account.FollowCount))

	return nil
}
