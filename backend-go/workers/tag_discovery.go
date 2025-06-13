package workers

import (
	"context"
	"fmt"
	"ftoolbox/config"
	"ftoolbox/fansly"
	"ftoolbox/models"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TagDiscoveryWorker struct {
	BaseWorker
	db       *gorm.DB
	client   *fansly.Client
	tagRegex *regexp.Regexp
	seedTags []string
}

func NewTagDiscoveryWorker(db *gorm.DB, cfg *config.Config) *TagDiscoveryWorker {
	interval := time.Duration(cfg.WorkerDiscoveryInterval) * time.Millisecond

	return &TagDiscoveryWorker{
		BaseWorker: NewBaseWorker("tag-discovery", interval),
		db:         db,
		client:     fansly.NewClient(cfg.FanslyAPIRateLimit),
		tagRegex:   regexp.MustCompile(`#(\w+)`),
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
	tagDetails, err := w.client.GetTag(tagToUse)
	if err != nil {
		// If tag not found on Fansly, mark it in database
		if err.Error() == "tag not found" {
			zap.L().Warn("Tag exists in database but not on Fansly, consider removing",
				zap.String("tag", tagToUse))
			// Optionally: Delete the tag or mark it as inactive
			// w.db.Delete(&models.Tag{}, "tag = ?", tagToUse)
		}
		return fmt.Errorf("failed to fetch tag details: %w", err)
	}

	// Fetch posts for this tag using its ID
	posts, err := w.client.GetPostsForTag(tagDetails.ID, 20, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %w", err)
	}

	// Extract and process tags
	discoveredTags := w.extractTags(posts)
	newTags := 0

	for _, tagName := range discoveredTags {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := w.processDiscoveredTag(tagName); err != nil {
			zap.L().Error("Failed to process discovered tag",
				zap.String("tag", tagName),
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
	// Get multiple tags that haven't been used recently, ordered by view count
	if err := w.db.Where("last_used_for_discovery IS NULL OR last_used_for_discovery < ?", sevenDaysAgo).
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

func (w *TagDiscoveryWorker) extractTags(posts []fansly.FanslyPost) []string {
	tagSet := make(map[string]bool)

	for _, post := range posts {
		// Extract from content using regex
		matches := w.tagRegex.FindAllStringSubmatch(post.Content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				tag := strings.ToLower(strings.TrimSpace(match[1]))
				if tag != "" {
					tagSet[tag] = true
				}
			}
		}
	}

	// Convert set to slice
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags
}

func (w *TagDiscoveryWorker) processDiscoveredTag(tagName string) error {
	// Check if tag already exists
	var existingTag models.Tag
	if err := w.db.Where("tag = ?", tagName).First(&existingTag).Error; err == nil {
		// Tag already exists
		return nil
	}

	// Fetch tag details from Fansly
	tagResponse, err := w.client.GetTag(tagName)
	if err != nil {
		// Tag might not exist on Fansly
		return fmt.Errorf("failed to fetch tag details: %w", err)
	}

	// Create new tag
	newTag := models.Tag{
		ID:              tagResponse.ID,
		Tag:             tagResponse.Tag,
		ViewCount:       tagResponse.ViewCount,
		FanslyCreatedAt: fansly.ParseFanslyTimestamp(tagResponse.CreatedAt),
	}

	if err := w.db.Create(&newTag).Error; err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	zap.L().Info("Discovered new tag",
		zap.String("tag", tagName),
		zap.Int64("viewCount", tagResponse.ViewCount))

	return nil
}
