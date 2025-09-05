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
	"gorm.io/gorm/clause"
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
			"amateur", "teen", "milf", "anal", "blonde", "brunette", "redhead",
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
	result, err := w.client.GetSuggestionsData(ctx, []string{tagDetails.MediaOfferSuggestionTag.ID}, "0", "0", 20, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %w", err)
	}

	// Extract and process tags from mediaOfferSuggestions
	discoveredTags := w.extractTagsFromSuggestions(result.MediaOfferSuggestions)
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

	// Update related tag relations from the suggestions
	if err := w.updateTagRelationsFromSuggestions(ctx, tagDetails.MediaOfferSuggestionTag.ID, result.MediaOfferSuggestions); err != nil {
		zap.L().Error("Failed to update tag relations", zap.Error(err))
	}

	// Purge old relation buckets beyond 14 days
	if err := w.purgeOldTagRelations(14); err != nil {
		zap.L().Error("Failed to purge old tag relations", zap.Error(err))
	}

	zap.L().Info("Tag discovery completed",
		zap.String("source_tag", tagToUse),
		zap.Int("discovered", len(discoveredTags)),
		zap.Int("new", newTags))

	tempCreatorWorker := NewCreatorUpdaterWorker(w.db, w.client)

	// Discover creators from the same tag
	if result.AggregationData != nil && result.AggregationData.Accounts != nil {
		if err := tempCreatorWorker.ProcessCreators(result.AggregationData.Accounts); err != nil {
			zap.L().Error("Failed to discover creators", zap.Error(err))
			// Don't return error, as tag discovery succeeded
		}
	}

	return nil
}

func (w *TagDiscoveryWorker) getTagForDiscovery() (string, error) {
	// First, try to get a tag from the database that hasn't been used recently
	hoursAgo := time.Now().Add(-3 * time.Hour)

	var tags []models.Tag
	// Get multiple tags that haven't been used recently and aren't deleted, ordered by rank
	if err := w.db.Where("(last_used_for_discovery IS NULL OR last_used_for_discovery < ?) AND is_deleted = ?", hoursAgo, false).
		Order("rank ASC").
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
		PostCount:       tag.PostCount,
		FanslyCreatedAt: fansly.ParseFanslyTimestamp(tag.CreatedAt),
	}

	if err := w.db.Create(&newTag).Error; err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	zap.L().Info("Discovered new tag",
		zap.String("tag", tag.Tag),
		zap.Int64("viewCount", tag.ViewCount),
		zap.Int64("postCount", tag.PostCount))

	return nil
}

// updateTagRelationsFromSuggestions records co-usage counts for all tags observed together per day.
// For each suggestion (post), it generates directed pairs for every tag to every other tag
// present on that post (A->B and B->A), counting each pair at most once per post.
func (w *TagDiscoveryWorker) updateTagRelationsFromSuggestions(ctx context.Context, sourceTagID string, suggestions []fansly.MediaOfferSuggestion) error {
	if sourceTagID == "" || len(suggestions) == 0 {
		return nil
	}

	// Bucket by date (UTC)
	bucketDate := time.Now().UTC().Truncate(24 * time.Hour)
	now := time.Now().UTC()

	// Aggregate per (source, related, date)
	type key struct{ source, related string }
	counts := make(map[key]int64)

	for _, s := range suggestions {
		// Build a unique set of tag IDs observed in this suggestion
		seen := make(map[string]struct{})
		for _, t := range s.PostTags {
			id := strings.TrimSpace(t.ID)
			if id == "" {
				continue
			}
			seen[id] = struct{}{}
		}
		// For all unique tags on this suggestion, generate directed pairs A->B (A != B)
		if len(seen) == 0 {
			continue
		}
		// Materialize keys to allow double loop without map iteration invalidation concerns
		ids := make([]string, 0, len(seen))
		for id := range seen {
			ids = append(ids, id)
		}
		for i := 0; i < len(ids); i++ {
			for j := 0; j < len(ids); j++ {
				if i == j {
					continue
				}
				counts[key{source: ids[i], related: ids[j]}]++
			}
		}
	}

	if len(counts) == 0 {
		return nil
	}

	// Prepare rows for upsert
	rows := make([]models.TagRelationDaily, 0, len(counts))
	for k, c := range counts {
		rows = append(rows, models.TagRelationDaily{
			TagID:        k.source,
			RelatedTagID: k.related,
			BucketDate:   bucketDate,
			CoCount:      c,
			LastSeenAt:   now,
		})
	}

	// Upsert with additive co_count and update last_seen_at
	if err := w.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "tag_id"}, {Name: "related_tag_id"}, {Name: "bucket_date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"co_count":     gorm.Expr("co_count + VALUES(co_count)"),
			"last_seen_at": gorm.Expr("VALUES(last_seen_at)"),
		}),
	}).Create(&rows).Error; err != nil {
		return fmt.Errorf("failed to upsert tag relations: %w", err)
	}

	// Optional: ensure related tags exist locally; skip here to avoid extra churn
	return nil
}

// purgeOldTagRelations removes data older than windowDays to cap storage
func (w *TagDiscoveryWorker) purgeOldTagRelations(windowDays int) error {
	if windowDays <= 0 {
		return nil
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -windowDays).Truncate(24 * time.Hour)
	tx := w.db.Where("bucket_date < ?", cutoff).Delete(&models.TagRelationDaily{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected > 0 {
		zap.L().Info("Purged old tag relations", zap.Int64("rows", tx.RowsAffected), zap.Time("cutoff", cutoff))
	}
	// Recalculate tag ranks/heat only if needed; not required for relations
	return nil
}
