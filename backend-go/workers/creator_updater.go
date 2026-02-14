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

const creatorUpdateBatchSize = 100

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

	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	var creators []models.Creator
	if err := w.db.Model(&models.Creator{}).
		Where("last_checked_at IS NULL OR last_checked_at < ?", twentyFourHoursAgo).
		Order("followers DESC").
		Limit(creatorUpdateBatchSize).
		Find(&creators).Error; err != nil {
		return fmt.Errorf("failed to fetch creators: %w", err)
	}

	if len(creators) == 0 {
		zap.L().Debug("No creators need updating")
		return nil
	}

	creatorIDs := make([]string, len(creators))
	for i, creator := range creators {
		creatorIDs[i] = creator.ID
	}

	zap.L().Info("Updating creators", zap.Int("count", len(creatorIDs)))

	accounts, err := w.client.GetAccountsWithContext(ctx, creatorIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch creator accounts: %w", err)
	}

	return w.processScheduledCreators(creators, accounts)
}

func (w *CreatorUpdaterWorker) processScheduledCreators(creators []models.Creator, accounts []fansly.FanslyAccount) error {
	if len(creators) == 0 {
		return nil
	}

	accountsByID := make(map[string]fansly.FanslyAccount, len(accounts))
	for _, account := range accounts {
		accountsByID[account.ID] = account
	}

	updatedCreators := 0
	missingCreators := 0

	for i := range creators {
		creator := creators[i]
		account, exists := accountsByID[creator.ID]
		if !exists {
			if err := w.markCreatorCheckedAfterMiss(&creator); err != nil {
				zap.L().Error("Failed to update creator after missing account lookup",
					zap.String("creator_id", creator.ID),
					zap.String("username", creator.Username),
					zap.Error(err))
				continue
			}
			missingCreators++
			continue
		}

		if err := w.updateCreator(&creator, &account); err != nil {
			zap.L().Error("Failed to update creator",
				zap.String("username", account.Username),
				zap.Error(err))
			continue
		}

		updatedCreators++
	}

	zap.L().Info("Creator updater run completed",
		zap.Int("requested", len(creators)),
		zap.Int("fetched", len(accounts)),
		zap.Int("updated", updatedCreators),
		zap.Int("missing", missingCreators))

	return nil
}

func (w *CreatorUpdaterWorker) markCreatorCheckedAfterMiss(creator *models.Creator) error {
	now := time.Now()
	updates := map[string]any{
		"last_checked_at": now,
		"updated_at":      now,
	}

	if err := w.db.Model(&models.Creator{}).Where("id = ?", creator.ID).Updates(updates).Error; err != nil {
		return err
	}

	return nil
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
