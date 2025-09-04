package handlers

import (
	"ftoolbox/fansly"
	"ftoolbox/models"
	"ftoolbox/utils"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TagHandler struct {
	db           *gorm.DB
	fanslyClient *fansly.Client
}

func NewTagHandler(db *gorm.DB, fanslyClient *fansly.Client) *TagHandler {
	return &TagHandler{
		db:           db,
		fanslyClient: fanslyClient,
	}
}

type HistoryPoint struct {
	ID              uint    `json:"id"`
	TagID           string  `json:"tagId"`
	ViewCount       int64   `json:"viewCount"`
	Change          int64   `json:"change"`
	PostCount       int64   `json:"postCount"`
	PostCountChange int64   `json:"postCountChange"`
	CreatedAt       int64   `json:"createdAt"`
	UpdatedAt       int64   `json:"updatedAt"`
	ChangePercent   float64 `json:"changePercent"`
}

type TagWithHistory struct {
	ID                   string         `json:"id"`
	Tag                  string         `json:"tag"`
	ViewCount            int64          `json:"viewCount"`
	PostCount            int64          `json:"postCount"`
	Rank                 *int           `json:"rank"`
	Heat                 float64        `json:"heat"`
	FanslyCreatedAt      *int64         `json:"fanslyCreatedAt"`
	LastCheckedAt        *int64         `json:"lastCheckedAt"`
	LastUsedForDiscovery *int64         `json:"lastUsedForDiscovery"`
	IsDeleted            bool           `json:"isDeleted"`
	DeletedDetectedAt    *int64         `json:"deletedDetectedAt"`
	CreatedAt            int64          `json:"createdAt"`
	UpdatedAt            int64          `json:"updatedAt"`
	History              []HistoryPoint `json:"history,omitempty"`
	TotalChange          int64          `json:"totalChange"`
}

func (h *TagHandler) GetTags(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	sortBy := c.Query("sortBy", "viewCount")
	sortOrder := c.Query("sortOrder", "desc")
	includeHistory := c.Query("includeHistory") == "true"
	historyStartDate := c.Query("historyStartDate")
	historyEndDate := c.Query("historyEndDate")
	tagsParam := c.Query("tags")

	// Parse tags parameter if provided
	var targetTags []string
	if tagsParam != "" {
		// Split by comma and trim whitespace
		targetTags = strings.Split(tagsParam, ",")
		for i := range targetTags {
			targetTags[i] = strings.TrimSpace(targetTags[i])
		}
		// Remove empty strings
		filtered := targetTags[:0]
		for _, tag := range targetTags {
			if tag != "" {
				filtered = append(filtered, tag)
			}
		}
		targetTags = filtered
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var tags []models.Tag
	query := h.db.Model(&models.Tag{})

	// If tags parameter is provided, filter by those specific tags
	if len(targetTags) > 0 {
		query = query.Where("tag IN ?", targetTags)
	} else if search != "" {
		// Only apply search filter if tags parameter is not provided
		query = query.Where("tag LIKE ?", "%"+search+"%")
	}

	// When sorting by rank, exclude tags without a rank from the base query
	if sortBy == "rank" {
		query = query.Where("rank IS NOT NULL")
	}

	var total int64
	query.Count(&total)

	// Handle sorting
	needsHistory := includeHistory || sortBy == "change"

	// Map frontend sortBy values to database columns
	columnMap := map[string]string{
		"viewCount": "view_count",
		"postCount": "post_count",
		"updatedAt": "updated_at",
		"tag":       "tag",
		"rank":      "rank",
		"heat":      "heat",
	}

	// For normal sorting
	if sortBy != "change" {
		dbColumn, ok := columnMap[sortBy]
		if !ok {
			dbColumn = "view_count" // default
		}

		orderClause := dbColumn + " " + sortOrder
		// Special case: when sorting by viewCount, treat deleted tags as having 0 view count
		if sortBy == "viewCount" {
			if sortOrder == "desc" {
				orderClause = "CASE WHEN is_deleted THEN 0 ELSE view_count END DESC, created_at DESC"
			} else {
				orderClause = "CASE WHEN is_deleted THEN 0 ELSE view_count END ASC, created_at ASC"
			}
		}
		query = query.Order(orderClause)

		// Only apply pagination if tags parameter is not provided
		if len(targetTags) == 0 {
			query = query.Limit(limit).Offset(offset)
		}

		if err := query.Find(&tags).Error; err != nil {
			zap.L().Error("Failed to fetch tags", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch tags"})
		}
	} else {
		// For change sorting, we need all tags to calculate aggregated values
		// But we can still apply search filter at DB level
		if err := query.Find(&tags).Error; err != nil {
			zap.L().Error("Failed to fetch tags", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch tags"})
		}
	}

	// If we need history, fetch it for each tag
	var tagsWithHistory []TagWithHistory
	if needsHistory {
		// Collect all tag IDs for batch loading
		tagIDs := make([]string, len(tags))
		tagMap := make(map[string]models.Tag)
		for i, tag := range tags {
			tagIDs[i] = tag.ID
			tagMap[tag.ID] = tag
		}

		// Process in batches to avoid too many placeholders
		const batchSize = 1000
		var allHistories []models.TagHistory

		for i := 0; i < len(tagIDs); i += batchSize {
			end := i + batchSize
			if end > len(tagIDs) {
				end = len(tagIDs)
			}
			batchIDs := tagIDs[i:end]

			// Build base query for this batch
			histQuery := h.db.Model(&models.TagHistory{}).
				Where("tag_id IN ?", batchIDs).
				Order("tag_id, created_at DESC")

			// Apply date filters if provided
			if historyStartDate != "" && historyEndDate != "" {
				// Try parsing as RFC3339 (ISO 8601) first, then fall back to date-only format
				startDate, err := time.Parse(time.RFC3339, historyStartDate)
				if err != nil {
					startDate, _ = time.Parse("2006-01-02", historyStartDate)
				}
				endDate, err := time.Parse(time.RFC3339, historyEndDate)
				if err != nil {
					endDate, _ = time.Parse("2006-01-02", historyEndDate)
				}
				histQuery = histQuery.Where("created_at >= ? AND created_at <= ?", startDate, endDate)
			}

			// Fetch histories for this batch
			var batchHistories []models.TagHistory
			if err := histQuery.Find(&batchHistories).Error; err != nil {
				zap.L().Error("Failed to fetch tag histories", zap.Error(err))
				return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch tag histories"})
			}
			allHistories = append(allHistories, batchHistories...)
		}

		// Group histories by tag ID
		historyByTag := make(map[string][]models.TagHistory)
		for _, hist := range allHistories {
			historyByTag[hist.TagID] = append(historyByTag[hist.TagID], hist)
		}

		// Process each tag with its history
		for _, tag := range tags {
			// Show 0 view count for deleted tags
			viewCount := tag.ViewCount
			if tag.IsDeleted {
				viewCount = 0
			}

			tagWithHist := TagWithHistory{
				ID:                   tag.ID,
				Tag:                  tag.Tag,
				ViewCount:            viewCount,
				PostCount:            tag.PostCount,
				Rank:                 tag.Rank,
				Heat:                 tag.Heat,
				FanslyCreatedAt:      ptr(timeToUnix(tag.FanslyCreatedAt)),
				LastCheckedAt:        timeToUnixPtr(tag.LastCheckedAt),
				LastUsedForDiscovery: timeToUnixPtr(tag.LastUsedForDiscovery),
				IsDeleted:            tag.IsDeleted,
				DeletedDetectedAt:    timeToUnixPtr(tag.DeletedDetectedAt),
				CreatedAt:            tag.CreatedAt.Unix(),
				UpdatedAt:            tag.UpdatedAt.Unix(),
			}

			history := historyByTag[tag.ID]

			// Calculate changes and convert to HistoryPoint
			historyPoints := make([]HistoryPoint, len(history))
			for i, point := range history {
				historyPoints[i] = HistoryPoint{
					ID:              point.ID,
					TagID:           point.TagID,
					ViewCount:       point.ViewCount,
					Change:          0,
					PostCount:       point.PostCount,
					PostCountChange: 0,
					CreatedAt:       point.CreatedAt.Unix(),
					UpdatedAt:       point.UpdatedAt.Unix(),
					ChangePercent:   0,
				}

				// Calculate change from previous point based on post count
				if i < len(history)-1 {
					previousPoint := history[i+1]
					// Still track view count change for reference
					viewChange := point.ViewCount - previousPoint.ViewCount
					historyPoints[i].Change = viewChange

					// Calculate post count change as primary metric
					postChange := point.PostCount - previousPoint.PostCount
					historyPoints[i].PostCountChange = postChange
					if previousPoint.PostCount > 0 {
						historyPoints[i].ChangePercent = float64(postChange) / float64(previousPoint.PostCount) * 100
					}
				}
			}

			// Calculate total change based on post count
			if len(history) > 0 {
				newest := history[0].PostCount
				oldest := history[len(history)-1].PostCount
				tagWithHist.TotalChange = newest - oldest
			}

			if includeHistory {
				tagWithHist.History = historyPoints
			}

			tagsWithHistory = append(tagsWithHistory, tagWithHist)
		}

		// Sort by change if requested
		if sortBy == "change" {
			// Use Go's efficient sort instead of bubble sort
			sort.Slice(tagsWithHistory, func(i, j int) bool {
				if sortOrder == "desc" {
					return tagsWithHistory[i].TotalChange > tagsWithHistory[j].TotalChange
				}
				return tagsWithHistory[i].TotalChange < tagsWithHistory[j].TotalChange
			})

			// Apply pagination after sorting only if tags parameter is not provided
			if len(targetTags) == 0 {
				start := offset
				end := offset + limit
				if start > len(tagsWithHistory) {
					tagsWithHistory = []TagWithHistory{}
				} else if end > len(tagsWithHistory) {
					tagsWithHistory = tagsWithHistory[start:]
				} else {
					tagsWithHistory = tagsWithHistory[start:end]
				}
			}
		}
	}

	// Return response with consistent format
	if needsHistory {
		return c.JSON(fiber.Map{
			"tags": tagsWithHistory,
			"pagination": fiber.Map{
				"page":       page,
				"limit":      limit,
				"totalCount": total,
				"totalPages": (total + int64(limit) - 1) / int64(limit),
			},
		})
	}

	// For non-history responses, we need to convert tags to proper format

	tagsData := make([]map[string]interface{}, len(tags))
	for i, tag := range tags {
		// Show 0 view count for deleted tags
		viewCount := tag.ViewCount
		if tag.IsDeleted {
			viewCount = 0
		}

		tagsData[i] = map[string]interface{}{
			"id":                   tag.ID,
			"tag":                  tag.Tag,
			"viewCount":            viewCount,
			"postCount":            tag.PostCount,
			"rank":                 tag.Rank,
			"heat":                 tag.Heat,
			"fanslyCreatedAt":      ptr(timeToUnix(tag.FanslyCreatedAt)),
			"lastCheckedAt":        timeToUnixPtr(tag.LastCheckedAt),
			"lastUsedForDiscovery": timeToUnixPtr(tag.LastUsedForDiscovery),
			"isDeleted":            tag.IsDeleted,
			"deletedDetectedAt":    timeToUnixPtr(tag.DeletedDetectedAt),
			"createdAt":            tag.CreatedAt.Unix(),
			"updatedAt":            tag.UpdatedAt.Unix(),
		}
	}

	return c.JSON(fiber.Map{
		"tags": tagsData,
		"pagination": fiber.Map{
			"page":       page,
			"limit":      limit,
			"totalCount": total,
			"totalPages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *TagHandler) GetTagStatistics(c *fiber.Ctx) error {
	// Get the most recent tag statistics from the database
	var stats models.TagStatistics
	if err := h.db.Order("calculated_at DESC").First(&stats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// No statistics exist yet, return zeros
			return c.JSON(fiber.Map{
				"totalViewCount":   0,
				"totalPostCount":   0,
				"change24h":        0,
				"changePercent24h": 0,
				"calculatedAt":     nil,
			})
		}
		zap.L().Error("Failed to fetch tag statistics", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch tag statistics",
		})
	}

	// Return the statistics with post count as primary metric
	return c.JSON(fiber.Map{
		"totalViewCount":   stats.TotalViewCount,
		"totalPostCount":   stats.TotalPostCount,
		"change24h":        stats.Change24h,
		"changePercent24h": stats.ChangePercent24h,
		"calculatedAt":     stats.CalculatedAt.Unix(),
	})
}

func (h *TagHandler) GetBannedTags(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	sortBy := c.Query("sortBy", "deletedDetectedAt")
	sortOrder := c.Query("sortOrder", "desc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var tags []models.Tag
	query := h.db.Model(&models.Tag{}).Where("is_deleted = ?", true)

	if search != "" {
		query = query.Where("tag LIKE ?", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	// Map frontend sortBy values to database columns
	columnMap := map[string]string{
		"tag":               "tag",
		"deletedDetectedAt": "deleted_detected_at",
		"viewCount":         "view_count",
		"postCount":         "post_count",
	}

	dbColumn, ok := columnMap[sortBy]
	if !ok {
		dbColumn = "deleted_detected_at" // default
	}

	orderClause := dbColumn + " " + sortOrder
	query = query.Order(orderClause)

	query = query.Limit(limit).Offset(offset)

	if err := query.Find(&tags).Error; err != nil {
		zap.L().Error("Failed to fetch banned tags", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch banned tags"})
	}

	// Calculate statistics for banned tags
	var stats struct {
		TotalBanned   int64 `json:"totalBanned"`
		BannedLast24h int64 `json:"bannedLast24h"`
		BannedLast7d  int64 `json:"bannedLast7d"`
		BannedLast30d int64 `json:"bannedLast30d"`
	}

	// Total banned tags
	h.db.Model(&models.Tag{}).Where("is_deleted = ?", true).Count(&stats.TotalBanned)

	// Banned in last 24 hours
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	h.db.Model(&models.Tag{}).Where("is_deleted = ? AND deleted_detected_at >= ?", true, oneDayAgo).Count(&stats.BannedLast24h)

	// Banned in last 7 days
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	h.db.Model(&models.Tag{}).Where("is_deleted = ? AND deleted_detected_at >= ?", true, sevenDaysAgo).Count(&stats.BannedLast7d)

	// Banned in last 30 days
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	h.db.Model(&models.Tag{}).Where("is_deleted = ? AND deleted_detected_at >= ?", true, thirtyDaysAgo).Count(&stats.BannedLast30d)

	return c.JSON(fiber.Map{
		"tags": tags,
		"pagination": fiber.Map{
			"page":       page,
			"limit":      limit,
			"totalCount": total,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
		},
		"statistics": stats,
	})
}

func (h *TagHandler) RequestTag(c *fiber.Ctx) error {
	var req struct {
		Tag string `json:"tag"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Tag == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Tag is required"})
	}

	// Check if tag already exists
	var existingTag models.Tag
	if err := h.db.Where("tag = ?", req.Tag).First(&existingTag).Error; err == nil {
		// Return existing tag like old backend
		return c.JSON(fiber.Map{
			"message": "Tag is already being tracked",
			"tag":     existingTag,
		})
	}

	// Immediately try to fetch tag data from Fansly
	fanslyTag, err := h.fanslyClient.GetTagWithContext(c.Context(), req.Tag)

	if err != nil || fanslyTag == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Tag not found on Fansly"})
	}

	// Insert tag into database
	newTag := models.Tag{
		ID:              fanslyTag.MediaOfferSuggestionTag.ID,
		Tag:             fanslyTag.MediaOfferSuggestionTag.Tag,
		ViewCount:       fanslyTag.MediaOfferSuggestionTag.ViewCount,
		PostCount:       fanslyTag.MediaOfferSuggestionTag.PostCount,
		FanslyCreatedAt: time.Unix(fanslyTag.MediaOfferSuggestionTag.CreatedAt/1000, 0),
		LastCheckedAt:   &[]time.Time{time.Now()}[0],
	}

	if err := h.db.Create(&newTag).Error; err != nil {
		zap.L().Error("Failed to create tag", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create tag"})
	}

	// Insert initial history record
	history := models.TagHistory{
		TagID:           newTag.ID,
		ViewCount:       newTag.ViewCount,
		Change:          0, // Initial entry has no change
		PostCount:       newTag.PostCount,
		PostCountChange: 0, // Initial entry has no change
	}

	if err := h.db.Create(&history).Error; err != nil {
		zap.L().Error("Failed to create tag history", zap.Error(err))
	}

	// Calculate ranks after adding the new tag
	if err := utils.CalculateTagRanks(h.db); err != nil {
		zap.L().Error("Failed to calculate ranks", zap.Error(err))
	}

	// Calculate heat scores after adding the new tag
	if err := utils.CalculateTagHeatScores(h.db); err != nil {
		zap.L().Error("Failed to calculate heat scores", zap.Error(err))
	}

	// Retrieve the tag again to get the calculated rank and heat
	var tagWithRank models.Tag
	if err := h.db.Where("id = ?", newTag.ID).First(&tagWithRank).Error; err != nil {
		zap.L().Error("Failed to retrieve tag with rank", zap.Error(err))
		// Return the tag without rank if retrieval fails
		return c.JSON(fiber.Map{
			"message": "Tag added successfully",
			"tag":     newTag,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Tag added successfully",
		"tag":     tagWithRank,
	})
}

// GetRelatedTags returns related tags based on co-usage observed in the last 14 days
func (h *TagHandler) GetRelatedTags(c *fiber.Ctx) error {
	// Parse inputs
	tagsParam := c.Query("tags", "")
	if strings.TrimSpace(tagsParam) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Query param 'tags' is required"})
	}
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	// Normalize and split tags
	parts := strings.Split(tagsParam, ",")
	inputs := make([]string, 0, len(parts))
	for _, t := range parts {
		v := strings.TrimSpace(strings.ToLower(t))
		if v != "" {
			inputs = append(inputs, v)
		}
	}
	if len(inputs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No valid tags provided"})
	}

	// Resolve to IDs
	var srcTags []models.Tag
	if err := h.db.Model(&models.Tag{}).Where("tag IN ?", inputs).Find(&srcTags).Error; err != nil {
		zap.L().Error("Failed to resolve tags", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to resolve tags"})
	}
	if len(srcTags) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No matching tags found"})
	}
	srcIDs := make([]string, 0, len(srcTags))
	for _, t := range srcTags {
		srcIDs = append(srcIDs, t.ID)
	}

	// Cutoff date for 14-day window (use date-only)
	cutoff := time.Now().UTC().AddDate(0, 0, -14).Truncate(24 * time.Hour)

	// Query related tags strictly by co-occurrence counts in the window
	// Filter out inputs themselves, deleted tags, and low-view tags (<5000)
	type row struct {
		ID    string `json:"id"`
		Tag   string `json:"tag"`
		Score int64  `json:"score"`
	}
	var rows []row

	// Build query using GORM
	// SELECT t.id, t.tag, SUM(tr.co_count) AS score
	// FROM tag_relations_daily tr
	// JOIN tags t ON t.id = tr.related_tag_id
	// WHERE tr.tag_id IN (?) AND tr.bucket_date >= ?
	//   AND t.is_deleted = FALSE AND t.view_count >= 5000
	//   AND tr.related_tag_id NOT IN (?)
	// GROUP BY t.id, t.tag
	// ORDER BY score DESC
	// LIMIT ?
	qb := h.db.Table("tag_relations_daily tr").
		Select("t.id as id, t.tag as tag, SUM(tr.co_count) as score").
		Joins("JOIN tags t ON t.id = tr.related_tag_id").
		Where("tr.tag_id IN ?", srcIDs).
		Where("tr.bucket_date >= ?", cutoff).
		Where("t.is_deleted = ?", false).
		Where("t.view_count >= ?", 5000).
		Where("tr.related_tag_id NOT IN ?", srcIDs).
		Group("t.id, t.tag").
		Order("score DESC").
		Limit(limit)

	if err := qb.Find(&rows).Error; err != nil {
		zap.L().Error("Failed to query related tags", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch related tags"})
	}

	// Build response
	resp := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, map[string]interface{}{
			"id":    r.ID,
			"tag":   r.Tag,
			"score": r.Score,
		})
	}

	return c.JSON(fiber.Map{
		"tags":         resp,
		"source":       "precomputed",
		"windowDays":   14,
		"minViewCount": 5000,
		"usedTagIds":   srcIDs,
	})
}

func timeToUnixPtr(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	return ptr(t.Unix())
}

func timeToUnix(t time.Time) int64 {
	return t.Unix()
}

func ptr[T any](v T) *T {
	return &v
}
