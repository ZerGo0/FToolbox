package handlers

import (
	"ftoolbox/fansly"
	"ftoolbox/models"
	"ftoolbox/utils"
	"sort"
	"strconv"
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
	ID            uint    `json:"id"`
	TagID         string  `json:"tagId"`
	ViewCount     int64   `json:"viewCount"`
	Change        int64   `json:"change"`
	CreatedAt     int64   `json:"createdAt"`
	UpdatedAt     int64   `json:"updatedAt"`
	ChangePercent float64 `json:"changePercent"`
}

type TagWithHistory struct {
	ID                   string         `json:"id"`
	Tag                  string         `json:"tag"`
	ViewCount            int64          `json:"viewCount"`
	Rank                 *int           `json:"rank"`
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

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var tags []models.Tag
	query := h.db.Model(&models.Tag{})

	if search != "" {
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
		"updatedAt": "updated_at",
		"tag":       "tag",
		"rank":      "rank",
	}

	// For normal sorting
	if sortBy != "change" {
		dbColumn, ok := columnMap[sortBy]
		if !ok {
			dbColumn = "view_count" // default
		}

		orderClause := dbColumn + " " + sortOrder
		// Special case: when sorting by viewCount desc, also sort by created_at desc as secondary
		if sortBy == "viewCount" && sortOrder == "desc" {
			orderClause = "view_count DESC, created_at DESC"
		}
		query = query.Order(orderClause).Limit(limit).Offset(offset)

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
			tagWithHist := TagWithHistory{
				ID:                   tag.ID,
				Tag:                  tag.Tag,
				ViewCount:            tag.ViewCount,
				Rank:                 tag.Rank,
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
					ID:            point.ID,
					TagID:         point.TagID,
					ViewCount:     point.ViewCount,
					Change:        0,
					CreatedAt:     point.CreatedAt.Unix(),
					UpdatedAt:     point.UpdatedAt.Unix(),
					ChangePercent: 0,
				}

				// Calculate change from previous point
				if i < len(history)-1 {
					previousPoint := history[i+1]
					change := point.ViewCount - previousPoint.ViewCount
					historyPoints[i].Change = change
					if previousPoint.ViewCount > 0 {
						historyPoints[i].ChangePercent = float64(change) / float64(previousPoint.ViewCount) * 100
					}
				}
			}

			// Calculate total change
			if len(history) > 0 {
				newest := history[0].ViewCount
				oldest := history[len(history)-1].ViewCount
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

			// Apply pagination after sorting
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
		tagsData[i] = map[string]interface{}{
			"id":                   tag.ID,
			"tag":                  tag.Tag,
			"viewCount":            tag.ViewCount,
			"rank":                 tag.Rank,
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

	// Return the statistics
	return c.JSON(fiber.Map{
		"totalViewCount":   stats.TotalViewCount,
		"change24h":        stats.Change24h,
		"changePercent24h": stats.ChangePercent24h,
		"calculatedAt":     stats.CalculatedAt.Unix(),
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
		FanslyCreatedAt: time.Unix(fanslyTag.MediaOfferSuggestionTag.CreatedAt/1000, 0),
		LastCheckedAt:   &[]time.Time{time.Now()}[0],
	}

	if err := h.db.Create(&newTag).Error; err != nil {
		zap.L().Error("Failed to create tag", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create tag"})
	}

	// Insert initial history record
	history := models.TagHistory{
		TagID:     newTag.ID,
		ViewCount: newTag.ViewCount,
		Change:    0, // Initial entry has no change
	}

	if err := h.db.Create(&history).Error; err != nil {
		zap.L().Error("Failed to create tag history", zap.Error(err))
	}

	// Calculate ranks after adding the new tag
	if err := utils.CalculateTagRanks(h.db); err != nil {
		zap.L().Error("Failed to calculate ranks", zap.Error(err))
	}

	// Retrieve the tag again to get the calculated rank
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
