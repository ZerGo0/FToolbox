package workers

import (
	"context"
	"errors"
	"fmt"
	"ftoolbox/config"
	"ftoolbox/fansly"
	"ftoolbox/models"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TagDiscoveryWorker struct {
	BaseWorker
	db       *gorm.DB
	client   *fansly.Client
	seedTags []string
}

func NewTagDiscoveryWorker(db *gorm.DB, cfg *config.Config, client *fansly.Client) *TagDiscoveryWorker {
	interval := time.Duration(cfg.WorkerDiscoveryInterval) * time.Millisecond

	return &TagDiscoveryWorker{
		BaseWorker: NewBaseWorker("tag-discovery", interval),
		db:         db,
		client:     client,
		seedTags: []string{
			"amateur", "teen", "milf", "anal", "asian", "latina", "ebony",
			"blonde", "brunette", "redhead", "bigboobs", "smalltits", "ass",
			"pussy", "blowjob", "cumshot", "creampie", "threesome", "lesbian",
			"fetish", "bdsm", "feet", "cosplay", "public", "outdoor", "shower",
			"masturbation", "toys", "lingerie", "stockings", "solo", "couple",
			"trans", "gay", "bisexual", "bbw", "mature", "hairy", "squirt",
			"dp", "gangbang", "orgy", "swingers", "cuckold", "femdom", "findom",
			"joi", "cei", "sph", "roleplay", "dirty", "naughty", "kinky",
		},
	}
}

func (w *TagDiscoveryWorker) Run(ctx context.Context) error {
	zap.L().Info("Running tag discovery")

	// Get a tag to use for discovery
	tagToUse, err := w.getTagForDiscovery()
	if err != nil {
		return fmt.Errorf("failed to get tag for discovery: %w", err)
	}

	if tagToUse == "" {
		zap.L().Debug("No suitable tag found for discovery")
		return nil
	}

	zap.L().Info("Discovering tags from", zap.String("source_tag", tagToUse))

	// Always update the last_used_for_discovery timestamp to prevent getting stuck
	// on the same tag if it fails
	defer func() {
		now := time.Now()
		if err := w.db.Model(&models.Tag{}).
			Where("tag = ?", tagToUse).
			Update("last_used_for_discovery", now).Error; err != nil {
			zap.L().Error("Failed to update last_used_for_discovery",
				zap.String("tag", tagToUse),
				zap.Error(err))
		}
	}()

	// First, get the tag details to get its ID
	tagDetails, err := w.client.GetTagWithContext(ctx, tagToUse)
	if err != nil {
		// If tag not found on Fansly, mark it as deleted
		if errors.Is(err, fansly.ErrTagNotFound) {
			zap.L().Info("Tag no longer exists on Fansly, marking as deleted",
				zap.String("tag", tagToUse))

			// Update the tag to mark it as deleted
			now := time.Now()
			updates := map[string]interface{}{
				"is_deleted":          true,
				"deleted_detected_at": &now,
				"updated_at":          now,
			}

			if updateErr := w.db.Model(&models.Tag{}).
				Where("tag = ?", tagToUse).
				Updates(updates).Error; updateErr != nil {
				zap.L().Error("Failed to mark tag as deleted",
					zap.String("tag", tagToUse),
					zap.Error(updateErr))
			}

			// Continue with discovery using another tag
			return nil
		}
		return fmt.Errorf("failed to fetch tag details: %w", err)
	}

	// Fetch posts for this tag using its ID
	result, err := w.client.GetPostsForTagWithPaginationAndContext(ctx, tagDetails.ID, 20, "0")
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %w", err)
	}

	// Extract and process tags from mediaOfferSuggestions
	discoveredTags := w.extractTagsFromSuggestions(result.Suggestions)
	newTags := 0

	for _, tag := range discoveredTags {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := w.processDiscoveredTag(tag); err != nil {
			zap.L().Error("Failed to process discovered tag",
				zap.String("tag", tag.Tag),
				zap.Error(err))
			continue
		} else {
			newTags++
		}
	}

	zap.L().Info("Tag discovery completed",
		zap.String("source_tag", tagToUse),
		zap.Int("discovered", len(discoveredTags)),
		zap.Int("new", newTags))

	return nil
}

func (w *TagDiscoveryWorker) getTagForDiscovery() (string, error) {
	// First, try to get a tag from the database that hasn't been used recently
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)

	var tags []models.Tag
	// Get multiple tags that haven't been used recently and aren't deleted, ordered by view count
	if err := w.db.Where("(last_used_for_discovery IS NULL OR last_used_for_discovery < ?) AND is_deleted = ?", sevenDaysAgo, false).
		Order("view_count DESC").
		Limit(10).
		Find(&tags).Error; err == nil && len(tags) > 0 {
		// Select a random tag from the top 10 to avoid always picking the same one
		idx := time.Now().Unix() % int64(len(tags))
		return tags[idx].Tag, nil
	}

	// If no suitable database tag, use a random seed tag
	if len(w.seedTags) > 0 {
		// Simple random selection using current time
		idx := time.Now().Unix() % int64(len(w.seedTags))
		return w.seedTags[idx], nil
	}

	return "", fmt.Errorf("no tags available for discovery")
}

// extractTagsFromSuggestions extracts unique tags from media offer suggestions
func (w *TagDiscoveryWorker) extractTagsFromSuggestions(suggestions []fansly.MediaOfferSuggestion) []fansly.FanslyTag {
	tagMap := make(map[string]fansly.FanslyTag)

	for _, suggestion := range suggestions {
		for _, tag := range suggestion.PostTags {
			// Use the tag name as key to ensure uniqueness
			tagName := strings.ToLower(strings.TrimSpace(tag.Tag))
			if tagName != "" {
				tagMap[tagName] = tag
			}
		}
	}

	// Convert map to slice
	tags := make([]fansly.FanslyTag, 0, len(tagMap))
	for _, tag := range tagMap {
		tags = append(tags, tag)
	}

	return tags
}

func (w *TagDiscoveryWorker) processDiscoveredTag(tag fansly.FanslyTag) error {
	// Check if tag already exists
	var existingTag models.Tag
	if err := w.db.Where("tag = ?", tag.Tag).First(&existingTag).Error; err == nil {
		// Tag already exists
		return nil
	}

	// Create new tag using the data we already have
	newTag := models.Tag{
		ID:              tag.ID,
		Tag:             tag.Tag,
		ViewCount:       tag.ViewCount,
		FanslyCreatedAt: fansly.ParseFanslyTimestamp(tag.CreatedAt),
	}

	if err := w.db.Create(&newTag).Error; err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	zap.L().Info("Discovered new tag",
		zap.String("tag", tag.Tag),
		zap.Int64("viewCount", tag.ViewCount))

	return nil
}
