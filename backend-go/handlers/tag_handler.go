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
	Ratio           float64 `json:"ratio"`
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
	Ratio                float64        `json:"ratio"`
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

type tagMetrics struct {
	ViewCount int64
	PostCount int64
	Ratio     float64
}

type relatedTagAggregate struct {
	ID          string  `json:"id"`
	Tag         string  `json:"tag"`
	RPostCount  int64   `json:"rPostCount"`
	NormSum     float64 `json:"normSum"`
	CoverageCnt int64   `json:"coverageCnt"`
}

type relatedTagScore struct {
	ID         string
	Tag        string
	NormAvg    float64
	Coverage   float64
	FinalScore float64
}

type tagSortOptions struct {
	By      string
	Order   string
	EndDate *time.Time
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
	sortBy := strings.ToLower(c.Query("sortBy", "rank"))
	sortOrder := strings.ToLower(c.Query("sortOrder", "asc"))
	includeHistory := c.Query("includeHistory") == "true"
	historyStartDate := c.Query("historyStartDate")
	historyEndDate := c.Query("historyEndDate")
	tagsParam := c.Query("tags")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	sortBy = sanitizeTagSortBy(sortBy)
	sortOrder = sanitizeSortOrder(sortOrder)

	offset := (page - 1) * limit
	startDate := parseHistoryDate(historyStartDate)
	endDate := parseHistoryDate(historyEndDate)
	targetTags, requestedTagsFilteredOut := parseRequestedTags(tagsParam)
	search, targetTags = resolveTagSearch(search, targetTags)

	var tags []models.Tag
	query := applyTagFilters(h.db.Model(&models.Tag{}), search, targetTags, requestedTagsFilteredOut).
		Where("rank IS NOT NULL")

	var total int64
	query.Count(&total)

	needsHistory := includeHistory
	query = h.applyTagSort(query, tagSortOptions{
		By:      sortBy,
		Order:   sortOrder,
		EndDate: endDate,
	})

	if len(targetTags) == 0 {
		query = query.Limit(limit).Offset(offset)
	}

	if err := query.Find(&tags).Error; err != nil {
		zap.L().Error("Failed to fetch tags", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch tags"})
	}

	tagIDs := collectTagIDs(tags)
	tagSnapshots, err := h.loadTagSnapshotsForRange(tagIDs, endDate)
	if err != nil {
		zap.L().Error("Failed to fetch tag snapshots", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch tag snapshots"})
	}

	if needsHistory {
		historyByTag, err := h.loadTagHistoryByTag(tagIDs, startDate, endDate)
		if err != nil {
			zap.L().Error("Failed to fetch tag histories", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch tag histories"})
		}

		return c.JSON(fiber.Map{
			"tags":       buildTagsWithHistory(tags, tagSnapshots, historyByTag, endDate),
			"pagination": buildPagination(page, limit, total),
		})
	}

	return c.JSON(fiber.Map{
		"tags":       buildTagData(tags, tagSnapshots, endDate),
		"pagination": buildPagination(page, limit, total),
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
	query := h.db.Model(&models.Tag{}).
		Where("is_deleted = ?", true).
		Where("tag NOT LIKE ?", "%+%")

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
	h.db.Model(&models.Tag{}).
		Where("is_deleted = ?", true).
		Where("tag NOT LIKE ?", "%+%").
		Count(&stats.TotalBanned)

	// Banned in last 24 hours
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	h.db.Model(&models.Tag{}).
		Where("is_deleted = ? AND deleted_detected_at >= ?", true, oneDayAgo).
		Where("tag NOT LIKE ?", "%+%").
		Count(&stats.BannedLast24h)

	// Banned in last 7 days
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	h.db.Model(&models.Tag{}).
		Where("is_deleted = ? AND deleted_detected_at >= ?", true, sevenDaysAgo).
		Where("tag NOT LIKE ?", "%+%").
		Count(&stats.BannedLast7d)

	// Banned in last 30 days
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	h.db.Model(&models.Tag{}).
		Where("is_deleted = ? AND deleted_detected_at >= ?", true, thirtyDaysAgo).
		Where("tag NOT LIKE ?", "%+%").
		Count(&stats.BannedLast30d)

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
	if validationError := validateRequestedTag(req.Tag); validationError != "" {
		return c.Status(400).JSON(fiber.Map{"error": validationError})
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
	if err := h.db.Model(&models.Tag{}).
		Where("tag IN ?", inputs).
		Where("tag NOT LIKE ?", "%+%").
		Find(&srcTags).Error; err != nil {
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
	var srows []relatedTagAggregate

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
		Where("t.tag NOT LIKE ?", "%+%").
		Where("t.view_count >= ?", minViewCount).
		Where("tr.related_tag_id NOT IN ?", srcIDs).
		Group("t.id, t.tag, t.post_count").
		Having("COUNT(DISTINCT tr.tag_id) >= ?", minCoverage)

	if err := qb.Find(&srows).Error; err != nil {
		zap.L().Error("Failed to query related tags (smart)", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch related tags"})
	}

	scoredRows := scoreRelatedTags(srows, len(srcIDs))

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
	return ptr(t.Unix())
}

func sanitizeTagSortBy(sortBy string) string {
	if sortBy == "ratio" {
		return "ratio"
	}
	return "rank"
}

func sanitizeSortOrder(sortOrder string) string {
	if sortOrder == "desc" {
		return "desc"
	}
	return "asc"
}

func parseRequestedTags(tagsParam string) ([]string, bool) {
	if tagsParam == "" {
		return nil, false
	}

	parts := strings.Split(tagsParam, ",")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" && !utils.TagNameHasPlus(tag) {
			filtered = append(filtered, tag)
		}
	}

	return filtered, len(filtered) == 0
}

func validateRequestedTag(tag string) string {
	if tag == "" {
		return "Tag is required"
	}
	if utils.TagNameHasPlus(tag) {
		return "Tags containing '+' are not supported"
	}
	return ""
}

func resolveTagSearch(search string, targetTags []string) (string, []string) {
	if len(targetTags) > 0 {
		return search, targetTags
	}

	if hashtags := extractHashtags(search); len(hashtags) > 0 {
		return "", hashtags
	}

	trimmed := strings.TrimSpace(search)
	if strings.HasPrefix(trimmed, "#") {
		return strings.TrimLeft(trimmed, "#"), targetTags
	}

	return search, targetTags
}

func applyTagFilters(
	query *gorm.DB,
	search string,
	targetTags []string,
	requestedTagsFilteredOut bool,
) *gorm.DB {
	query = query.Where("tag NOT LIKE ?", "%+%")

	if requestedTagsFilteredOut {
		return query.Where("1 = 0")
	}
	if len(targetTags) > 0 {
		return query.Where("tag IN ?", targetTags)
	}
	if search != "" {
		return query.Where("tag LIKE ?", "%"+search+"%")
	}

	return query
}

func collectTagIDs(tags []models.Tag) []string {
	tagIDs := make([]string, len(tags))
	for i, tag := range tags {
		tagIDs[i] = tag.ID
	}

	return tagIDs
}

func buildTagMetrics(tag models.Tag, snapshots map[string]models.TagHistory, endDate *time.Time) tagMetrics {
	metrics := tagMetrics{
		ViewCount: tag.ViewCount,
		PostCount: tag.PostCount,
	}

	if snapshot, ok := snapshots[tag.ID]; ok {
		metrics.ViewCount = snapshot.ViewCount
		metrics.PostCount = snapshot.PostCount
	}
	if tagDeletedByRangeEnd(tag, endDate) {
		metrics.ViewCount = 0
	}

	metrics.Ratio = utils.CalculateRatio(metrics.ViewCount, metrics.PostCount)
	return metrics
}

func buildTagHistoryPoints(history []models.TagHistory) []HistoryPoint {
	historyPoints := make([]HistoryPoint, len(history))
	for i, point := range history {
		historyPoint := HistoryPoint{
			ID:              point.ID,
			TagID:           point.TagID,
			ViewCount:       point.ViewCount,
			Change:          0,
			PostCount:       point.PostCount,
			Ratio:           utils.CalculateRatio(point.ViewCount, point.PostCount),
			PostCountChange: 0,
			CreatedAt:       point.CreatedAt.Unix(),
			UpdatedAt:       point.UpdatedAt.Unix(),
			ChangePercent:   0,
		}

		if i < len(history)-1 {
			previousPoint := history[i+1]
			viewChange := point.ViewCount - previousPoint.ViewCount
			historyPoint.Change = viewChange
			historyPoint.PostCountChange = point.PostCount - previousPoint.PostCount
			if previousPoint.ViewCount > 0 {
				historyPoint.ChangePercent = float64(viewChange) / float64(previousPoint.ViewCount) * 100
			}
		}

		historyPoints[i] = historyPoint
	}

	return historyPoints
}

func buildTagWithHistory(
	tag models.Tag,
	snapshots map[string]models.TagHistory,
	history []models.TagHistory,
	endDate *time.Time,
) TagWithHistory {
	metrics := buildTagMetrics(tag, snapshots, endDate)
	tagWithHistory := TagWithHistory{
		ID:                   tag.ID,
		Tag:                  tag.Tag,
		ViewCount:            metrics.ViewCount,
		PostCount:            metrics.PostCount,
		Ratio:                metrics.Ratio,
		Rank:                 tag.Rank,
		Heat:                 0,
		FanslyCreatedAt:      ptr(timeToUnix(tag.FanslyCreatedAt)),
		LastCheckedAt:        timeToUnixPtr(tag.LastCheckedAt),
		LastUsedForDiscovery: timeToUnixPtr(tag.LastUsedForDiscovery),
		IsDeleted:            tag.IsDeleted,
		DeletedDetectedAt:    timeToUnixPtr(tag.DeletedDetectedAt),
		CreatedAt:            tag.CreatedAt.Unix(),
		UpdatedAt:            tag.UpdatedAt.Unix(),
		History:              buildTagHistoryPoints(history),
	}

	if len(history) > 0 {
		tagWithHistory.TotalChange = history[0].ViewCount - history[len(history)-1].ViewCount
	}

	return tagWithHistory
}

func buildTagsWithHistory(
	tags []models.Tag,
	snapshots map[string]models.TagHistory,
	historyByTag map[string][]models.TagHistory,
	endDate *time.Time,
) []TagWithHistory {
	tagsWithHistory := make([]TagWithHistory, 0, len(tags))
	for _, tag := range tags {
		tagsWithHistory = append(
			tagsWithHistory,
			buildTagWithHistory(tag, snapshots, historyByTag[tag.ID], endDate),
		)
	}

	return tagsWithHistory
}

func buildTagData(
	tags []models.Tag,
	snapshots map[string]models.TagHistory,
	endDate *time.Time,
) []map[string]any {
	tagsData := make([]map[string]any, len(tags))
	for i, tag := range tags {
		metrics := buildTagMetrics(tag, snapshots, endDate)
		tagsData[i] = map[string]any{
			"id":                   tag.ID,
			"tag":                  tag.Tag,
			"viewCount":            metrics.ViewCount,
			"postCount":            metrics.PostCount,
			"ratio":                metrics.Ratio,
			"rank":                 tag.Rank,
			"heat":                 0,
			"fanslyCreatedAt":      ptr(timeToUnix(tag.FanslyCreatedAt)),
			"lastCheckedAt":        timeToUnixPtr(tag.LastCheckedAt),
			"lastUsedForDiscovery": timeToUnixPtr(tag.LastUsedForDiscovery),
			"isDeleted":            tag.IsDeleted,
			"deletedDetectedAt":    timeToUnixPtr(tag.DeletedDetectedAt),
			"createdAt":            tag.CreatedAt.Unix(),
			"updatedAt":            tag.UpdatedAt.Unix(),
		}
	}

	return tagsData
}

func (h *TagHandler) applyTagSort(
	query *gorm.DB,
	sortOptions tagSortOptions,
) *gorm.DB {
	if sortOptions.By != "ratio" {
		return query.Order("rank " + sortOptions.Order)
	}

	orderClause := currentTagRatioOrderClause()
	if sortOptions.EndDate != nil {
		query = query.Joins(latestTagSnapshotJoinForTagsClause(), *sortOptions.EndDate)
		orderClause = snapshotTagRatioOrderClause()
	}

	return query.Order(orderClause + " " + sortOptions.Order).Order("rank ASC")
}

func (h *TagHandler) loadTagSnapshotsForRange(
	tagIDs []string,
	endDate *time.Time,
) (map[string]models.TagHistory, error) {
	if endDate == nil {
		return map[string]models.TagHistory{}, nil
	}

	return loadTagSnapshots(h.db, tagIDs, *endDate)
}

func (h *TagHandler) loadTagHistoryByTag(
	tagIDs []string,
	startDate *time.Time,
	endDate *time.Time,
) (map[string][]models.TagHistory, error) {
	historyByTag := make(map[string][]models.TagHistory)
	if len(tagIDs) == 0 {
		return historyByTag, nil
	}

	const batchSize = 1000
	for i := 0; i < len(tagIDs); i += batchSize {
		end := min(i+batchSize, len(tagIDs))
		batchIDs := tagIDs[i:end]

		histQuery := h.db.Model(&models.TagHistory{}).
			Where("tag_id IN ?", batchIDs).
			Order("tag_id, created_at DESC")

		if startDate != nil {
			histQuery = histQuery.Where("created_at >= ?", *startDate)
		}
		if endDate != nil {
			histQuery = histQuery.Where("created_at <= ?", *endDate)
		}

		var batchHistories []models.TagHistory
		if err := histQuery.Find(&batchHistories).Error; err != nil {
			return nil, err
		}

		for _, history := range batchHistories {
			historyByTag[history.TagID] = append(historyByTag[history.TagID], history)
		}
	}

	return historyByTag, nil
}

func loadTagSnapshots(db *gorm.DB, tagIDs []string, endDate time.Time) (map[string]models.TagHistory, error) {
	if len(tagIDs) == 0 {
		return map[string]models.TagHistory{}, nil
	}

	var snapshots []models.TagHistory
	if err := db.Table("tags AS t").
		Select(
			"tag_snapshots.id, tag_snapshots.tag_id, tag_snapshots.view_count, tag_snapshots.change, "+
				"tag_snapshots.post_count, tag_snapshots.post_count_change, tag_snapshots.created_at, tag_snapshots.updated_at",
		).
		Joins(latestTagSnapshotJoinForSnapshotLoadClause(), endDate).
		Where("t.id IN ?", tagIDs).
		Where("tag_snapshots.id IS NOT NULL").
		Scan(&snapshots).Error; err != nil {
		return nil, err
	}

	snapshotByTag := make(map[string]models.TagHistory, len(snapshots))
	for _, snapshot := range snapshots {
		snapshotByTag[snapshot.TagID] = snapshot
	}

	return snapshotByTag, nil
}

func currentTagRatioOrderClause() string {
	return "(CASE WHEN tags.post_count > 0 THEN " +
		"CAST(tags.view_count AS DECIMAL(30,10)) / tags.post_count" +
		" ELSE 0 END)"
}

func snapshotTagRatioOrderClause() string {
	return "(CASE WHEN COALESCE(tag_snapshots.post_count, tags.post_count) > 0 THEN " +
		"CAST(COALESCE(tag_snapshots.view_count, tags.view_count) AS DECIMAL(30,10)) / " +
		"COALESCE(tag_snapshots.post_count, tags.post_count)" +
		" ELSE 0 END)"
}

func latestTagSnapshotJoinForTagsClause() string {
	return "LEFT JOIN tag_history AS tag_snapshots ON tag_snapshots.id = (" +
		"SELECT th.id FROM tag_history AS th " +
		"WHERE th.tag_id = tags.id AND th.created_at <= ? " +
		"ORDER BY th.created_at DESC, th.id DESC LIMIT 1)"
}

func latestTagSnapshotJoinForSnapshotLoadClause() string {
	return "LEFT JOIN tag_history AS tag_snapshots ON tag_snapshots.id = (" +
		"SELECT th.id FROM tag_history AS th " +
		"WHERE th.tag_id = t.id AND th.created_at <= ? " +
		"ORDER BY th.created_at DESC, th.id DESC LIMIT 1)"
}

func scoreRelatedTags(rows []relatedTagAggregate, inputCount int) []relatedTagScore {
	numInputs := float64(inputCount)
	scoredRows := make([]relatedTagScore, 0, len(rows))

	for _, row := range rows {
		normAvg := 0.0
		coverage := 0.0
		if numInputs > 0 {
			normAvg = row.NormSum / numInputs
			coverage = float64(row.CoverageCnt) / numInputs
		}

		postCount := float64(row.RPostCount)
		if postCount < 0 {
			postCount = 0
		}
		if postCount > 50000 {
			postCount = 50000
		}

		popBoost := 1.0
		logValue := math.Log1p(postCount)
		if logValue > 0 {
			popBoost = math.Pow(logValue, 0.2)
		}

		scoredRows = append(scoredRows, relatedTagScore{
			ID:         row.ID,
			Tag:        row.Tag,
			NormAvg:    normAvg,
			Coverage:   coverage,
			FinalScore: normAvg * coverage * popBoost,
		})
	}

	return scoredRows
}

func tagDeletedByRangeEnd(tag models.Tag, endDate *time.Time) bool {
	if !tag.IsDeleted {
		return false
	}
	if endDate == nil || tag.DeletedDetectedAt == nil {
		return true
	}
	return !tag.DeletedDetectedAt.After(*endDate)
}

func timeToUnix(t time.Time) int64 {
	return t.Unix()
}

//go:fix inline
func ptr[T any](v T) *T {
	return &v
}
