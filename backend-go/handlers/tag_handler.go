package handlers

import (
	"ftoolbox/fansly"
	"ftoolbox/models"
	"ftoolbox/utils"
	"math"
	"regexp"
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

func extractHashtags(q string) []string {
	if strings.TrimSpace(q) == "" {
		return nil
	}
	r := regexp.MustCompile(`#([\p{L}\p{N}_-]+)`) // letters, numbers, underscore, dash
	matches := r.FindAllStringSubmatch(q, -1)
	if len(matches) == 0 {
		if strings.HasPrefix(strings.TrimSpace(q), "#") {
			v := strings.TrimLeft(strings.TrimSpace(q), "#")
			if v != "" {
				return []string{v}
			}
		}
		return nil
	}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 && m[1] != "" {
			out = append(out, m[1])
		}
	}
	return out
}

func (h *TagHandler) GetTags(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	sortOrder := strings.ToLower(c.Query("sortOrder", "asc"))
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

	if len(targetTags) == 0 {
		if hs := extractHashtags(search); len(hs) > 0 {
			targetTags = hs
			search = ""
		} else if strings.HasPrefix(strings.TrimSpace(search), "#") {
			search = strings.TrimLeft(strings.TrimSpace(search), "#")
		}
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
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

	query = query.Where("rank IS NOT NULL")

	var total int64
	query.Count(&total)

	// Handle sorting
	needsHistory := includeHistory

	orderClause := "rank " + sortOrder
	query = query.Order(orderClause)

	if len(targetTags) == 0 {
		query = query.Limit(limit).Offset(offset)
	}

	if err := query.Find(&tags).Error; err != nil {
		zap.L().Error("Failed to fetch tags", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch tags"})
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
			end := min(i+batchSize, len(tagIDs))
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
				Heat:                 0,
				FanslyCreatedAt:      new(timeToUnix(tag.FanslyCreatedAt)),
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

				// Calculate change from previous point based on view count
				if i < len(history)-1 {
					previousPoint := history[i+1]
					// Calculate view count change as primary metric
					viewChange := point.ViewCount - previousPoint.ViewCount
					historyPoints[i].Change = viewChange

					// Still track post count change for reference
					postChange := point.PostCount - previousPoint.PostCount
					historyPoints[i].PostCountChange = postChange
					if previousPoint.ViewCount > 0 {
						historyPoints[i].ChangePercent = float64(viewChange) / float64(previousPoint.ViewCount) * 100
					}
				}
			}

			// Calculate total change based on view count
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

	tagsData := make([]map[string]any, len(tags))
	for i, tag := range tags {
		// Show 0 view count for deleted tags
		viewCount := tag.ViewCount
		if tag.IsDeleted {
			viewCount = 0
		}

		tagsData[i] = map[string]any{
			"id":                   tag.ID,
			"tag":                  tag.Tag,
			"viewCount":            viewCount,
			"postCount":            tag.PostCount,
			"rank":                 tag.Rank,
			"heat":                 0,
			"fanslyCreatedAt":      new(timeToUnix(tag.FanslyCreatedAt)),
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
				"totalViewCount":       0,
				"totalPostCount":       0,
				"change24h":            0,
				"changePercent24h":     0,
				"postChange24h":        0,
				"postChangePercent24h": 0,
				"calculatedAt":         nil,
			})
		}
		zap.L().Error("Failed to fetch tag statistics", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch tag statistics",
		})
	}

	return c.JSON(fiber.Map{
		"totalViewCount":       stats.TotalViewCount,
		"totalPostCount":       stats.TotalPostCount,
		"change24h":            stats.Change24h,
		"changePercent24h":     stats.ChangePercent24h,
		"postChange24h":        stats.PostChange24h,
		"postChangePercent24h": stats.PostChangePercent24h,
		"calculatedAt":         stats.CalculatedAt.Unix(),
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

	if hs := extractHashtags(search); len(hs) > 0 {
		query = query.Where("tag IN ?", hs)
	} else if strings.HasPrefix(strings.TrimSpace(search), "#") {
		v := strings.TrimLeft(strings.TrimSpace(search), "#")
		if v != "" {
			query = query.Where("tag LIKE ?", "%"+v+"%")
		}
	} else if search != "" {
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

	// Force heat to 0 in responses
	for i := range tags {
		tags[i].Heat = 0
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

	// Ensure heat is 0 in response
	tagWithRank.Heat = 0

	return c.JSON(fiber.Map{
		"message": "Tag added successfully",
		"tag":     tagWithRank,
	})
}

// GetRelatedTags returns related tags based on co-usage observed in a recent window
// Smart scoring: per-source normalization, coverage weighting, light popularity shaping
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

	// Legacy "popular" mode removed; force smart mode
	mode := "smart"

	windowDays, _ := strconv.Atoi(c.Query("windowDays", "14"))
	if windowDays < 7 {
		windowDays = 14 // clamp to sane defaults
	}
	if windowDays > 30 {
		windowDays = 30
	}

	minViewCount, _ := strconv.Atoi(c.Query("minViewCount", "5000"))
	if minViewCount < 0 {
		minViewCount = 0
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

	// Cutoff date for window (use date-only)
	cutoff := time.Now().UTC().AddDate(0, 0, -windowDays).Truncate(24 * time.Hour)

	// Legacy "popular" mode branch removed

	// Smart mode: per-source normalization + coverage weighting
	// minCoverage: default to ceil(40% of inputs), bounded to [1, len(inputs)]
	minCoverageDefault := max(int(math.Ceil(0.4*float64(len(srcIDs)))), 1)
	if minCoverageDefault > len(srcIDs) {
		minCoverageDefault = len(srcIDs)
	}
	minCoverage, _ := strconv.Atoi(c.Query("minCoverage", strconv.Itoa(minCoverageDefault)))
	if minCoverage < 1 {
		minCoverage = 1
	}
	if minCoverage > len(srcIDs) {
		minCoverage = len(srcIDs)
	}

	// Query base aggregates; compute popularity shaping and final score in Go for portability
	type smartRow struct {
		ID          string  `json:"id"`
		Tag         string  `json:"tag"`
		RPostCount  int64   `json:"rPostCount"`
		NormSum     float64 `json:"normSum"`
		CoverageCnt int64   `json:"coverageCnt"`
	}
	var srows []smartRow

	// Use CAST to float to ensure floating division across dialects
	selectExpr := strings.Join([]string{
		"t.id as id",
		"t.tag as tag",
		"t.post_count as r_post_count",
		"SUM(CAST(tr.co_count AS DOUBLE) / NULLIF(ts.post_count, 0)) as norm_sum",
		"COUNT(DISTINCT tr.tag_id) as coverage_cnt",
	}, ", ")

	qb := h.db.Table("tag_relations_daily tr").
		Select(selectExpr).
		Joins("JOIN tags t ON t.id = tr.related_tag_id").
		Joins("JOIN tags ts ON ts.id = tr.tag_id").
		Where("tr.tag_id IN ?", srcIDs).
		Where("tr.bucket_date >= ?", cutoff).
		Where("t.is_deleted = ?", false).
		Where("t.view_count >= ?", minViewCount).
		Where("tr.related_tag_id NOT IN ?", srcIDs).
		Group("t.id, t.tag, t.post_count").
		Having("COUNT(DISTINCT tr.tag_id) >= ?", minCoverage)

	if err := qb.Find(&srows).Error; err != nil {
		zap.L().Error("Failed to query related tags (smart)", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch related tags"})
	}

	numInputs := float64(len(srcIDs))
	// Compute final scores and sort
	type scored struct {
		ID         string
		Tag        string
		NormAvg    float64
		Coverage   float64
		FinalScore float64
	}
	scoredRows := make([]scored, 0, len(srows))
	for _, r := range srows {
		// Safety for division
		na := 0.0
		if numInputs > 0 {
			na = r.NormSum / numInputs
		}
		cov := 0.0
		if numInputs > 0 {
			cov = float64(r.CoverageCnt) / numInputs
		}
		// Popularity shaping: gentle boost to avoid ultra-rare dominating
		pc := float64(r.RPostCount)
		if pc < 0 {
			pc = 0
		}
		if pc > 50000 {
			pc = 50000
		}
		popBoost := 1.0
		// Avoid NaN if log1p(0) -> 0; keep boost >= 1 slightly for non-zero
		logv := math.Log1p(pc)
		if logv > 0 {
			popBoost = math.Pow(logv, 0.2)
		}
		final := na * cov * popBoost
		scoredRows = append(scoredRows, scored{
			ID:         r.ID,
			Tag:        r.Tag,
			NormAvg:    na,
			Coverage:   cov,
			FinalScore: final,
		})
	}

	sort.Slice(scoredRows, func(i, j int) bool {
		return scoredRows[i].FinalScore > scoredRows[j].FinalScore
	})

	if limit > len(scoredRows) {
		limit = len(scoredRows)
	}
	out := scoredRows[:limit]

	resp := make([]map[string]any, 0, len(out))
	for _, r := range out {
		resp = append(resp, map[string]any{
			"id":         r.ID,
			"tag":        r.Tag,
			"normScore":  r.NormAvg,
			"coverage":   r.Coverage,
			"finalScore": r.FinalScore,
			"score":      r.FinalScore, // Back-compat: keep 'score'
		})
	}

	return c.JSON(fiber.Map{
		"tags":         resp,
		"source":       "computed",
		"mode":         mode,
		"windowDays":   windowDays,
		"minViewCount": minViewCount,
		"minCoverage":  minCoverage,
		"usedTagIds":   srcIDs,
	})
}

func timeToUnixPtr(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	return new(t.Unix())
}

func timeToUnix(t time.Time) int64 {
	return t.Unix()
}

//go:fix inline
func ptr[T any](v T) *T {
	return new(v)
}
