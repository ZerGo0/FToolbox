package handlers

import (
	"ftoolbox/fansly"
	"ftoolbox/models"
	"ftoolbox/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CreatorHandler struct {
	db           *gorm.DB
	fanslyClient *fansly.Client
}

func NewCreatorHandler(db *gorm.DB, fanslyClient *fansly.Client) *CreatorHandler {
	return &CreatorHandler{
		db:           db,
		fanslyClient: fanslyClient,
	}
}

type CreatorHistoryPoint struct {
	ID         uint   `json:"id"`
	CreatorID  string `json:"creatorId"`
	MediaLikes int64  `json:"mediaLikes"`
	PostLikes  int64  `json:"postLikes"`
	Followers  int64  `json:"followers"`
	ImageCount int64  `json:"imageCount"`
	VideoCount int64  `json:"videoCount"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

type CreatorWithHistory struct {
	ID                string                `json:"id"`
	Username          string                `json:"username"`
	DisplayName       *string               `json:"displayName"`
	MediaLikes        int64                 `json:"mediaLikes"`
	PostLikes         int64                 `json:"postLikes"`
	Followers         int64                 `json:"followers"`
	ImageCount        int64                 `json:"imageCount"`
	VideoCount        int64                 `json:"videoCount"`
	Rank              *int                  `json:"rank"`
	LastCheckedAt     *int64                `json:"lastCheckedAt"`
	IsDeleted         bool                  `json:"isDeleted"`
	DeletedDetectedAt *int64                `json:"deletedDetectedAt"`
	CreatedAt         int64                 `json:"createdAt"`
	UpdatedAt         int64                 `json:"updatedAt"`
	History           []CreatorHistoryPoint `json:"history,omitempty"`
}

type creatorMetrics struct {
	MediaLikes int64
	PostLikes  int64
	Followers  int64
	ImageCount int64
	VideoCount int64
}

func (h *CreatorHandler) GetCreators(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	sortOrder := strings.ToLower(c.Query("sortOrder", "asc"))
	includeHistory := c.Query("includeHistory") == "true"
	historyStartDate := c.Query("historyStartDate")
	historyEndDate := c.Query("historyEndDate")

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
	startDate := parseHistoryDate(historyStartDate)
	endDate := parseHistoryDate(historyEndDate)

	var creators []models.Creator
	query := applyCreatorSearch(h.db.Model(&models.Creator{}), search).Where("rank IS NOT NULL")

	var total int64
	query.Count(&total)

	needsHistory := includeHistory
	query = query.Order("rank " + sortOrder).Limit(limit).Offset(offset)

	if err := query.Find(&creators).Error; err != nil {
		zap.L().Error("Failed to fetch creators", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch creators"})
	}

	creatorIDs := collectCreatorIDs(creators)
	creatorSnapshots, err := h.loadCreatorSnapshotsForRange(creatorIDs, endDate)
	if err != nil {
		zap.L().Error("Failed to fetch creator snapshots", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch creator snapshots"})
	}

	if needsHistory {
		historyByCreator, err := h.loadCreatorHistoryByCreator(creatorIDs, startDate, endDate)
		if err != nil {
			zap.L().Error("Failed to fetch creator histories", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch creator histories"})
		}

		return c.JSON(fiber.Map{
			"creators":   buildCreatorsWithHistory(creators, creatorSnapshots, historyByCreator),
			"pagination": buildPagination(page, limit, total),
		})
	}

	return c.JSON(fiber.Map{
		"creators":   buildCreatorData(creators, creatorSnapshots),
		"pagination": buildPagination(page, limit, total),
	})
}

func (h *CreatorHandler) GetCreatorStatistics(c *fiber.Ctx) error {
	// Get the most recent creator statistics from the database
	var stats models.CreatorStatistics
	if err := h.db.Order("calculated_at DESC").First(&stats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// No statistics exist yet, return zeros
			return c.JSON(fiber.Map{
				"totalFollowers":             0,
				"followersChange24h":         0,
				"followersChangePercent24h":  0,
				"totalMediaLikes":            0,
				"mediaLikesChange24h":        0,
				"mediaLikesChangePercent24h": 0,
				"totalPostLikes":             0,
				"postLikesChange24h":         0,
				"postLikesChangePercent24h":  0,
				"calculatedAt":               nil,
			})
		}
		zap.L().Error("Failed to fetch creator statistics", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch creator statistics",
		})
	}

	// Return the statistics
	return c.JSON(fiber.Map{
		"totalFollowers":             stats.TotalFollowers,
		"followersChange24h":         stats.FollowersChange24h,
		"followersChangePercent24h":  stats.FollowersChangePercent24h,
		"totalMediaLikes":            stats.TotalMediaLikes,
		"mediaLikesChange24h":        stats.MediaLikesChange24h,
		"mediaLikesChangePercent24h": stats.MediaLikesChangePercent24h,
		"totalPostLikes":             stats.TotalPostLikes,
		"postLikesChange24h":         stats.PostLikesChange24h,
		"postLikesChangePercent24h":  stats.PostLikesChangePercent24h,
		"calculatedAt":               stats.CalculatedAt.Unix(),
	})
}

func (h *CreatorHandler) RequestCreator(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Username == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Username is required"})
	}

	// Check if creator already exists
	var existingCreator models.Creator
	if err := h.db.Where("username = ?", req.Username).First(&existingCreator).Error; err == nil {
		// Return existing creator
		return c.JSON(fiber.Map{
			"message": "Creator is already being tracked",
			"creator": existingCreator,
		})
	}

	// Immediately try to fetch creator data from Fansly
	fanslyAccount, err := h.fanslyClient.GetAccountByUsername(c.Context(), req.Username)

	if err != nil || fanslyAccount == nil {
		if err != nil {
			zap.L().Error("Failed to fetch creator data from Fansly", zap.Error(err))
		} else {
			zap.L().Warn("DEBUGDEBUGDEBUGDEBUG Creator not found on Fansly")
		}
		return c.Status(404).JSON(fiber.Map{"error": "Creator not found on Fansly"})
	}

	// Insert creator into database
	newCreator := models.Creator{
		ID:            fanslyAccount.ID,
		Username:      fanslyAccount.Username,
		DisplayName:   &fanslyAccount.DisplayName,
		MediaLikes:    fanslyAccount.AccountMediaLikes,
		PostLikes:     fanslyAccount.PostLikes,
		Followers:     fanslyAccount.FollowCount,
		ImageCount:    fanslyAccount.TimelineStats.ImageCount,
		VideoCount:    fanslyAccount.TimelineStats.VideoCount,
		LastCheckedAt: &[]time.Time{time.Now()}[0],
	}

	if err := h.db.Create(&newCreator).Error; err != nil {
		zap.L().Error("Failed to create creator", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create creator"})
	}

	// Insert initial history record
	history := models.CreatorHistory{
		CreatorID:  newCreator.ID,
		MediaLikes: newCreator.MediaLikes,
		PostLikes:  newCreator.PostLikes,
		Followers:  newCreator.Followers,
		ImageCount: newCreator.ImageCount,
		VideoCount: newCreator.VideoCount,
	}

	if err := h.db.Create(&history).Error; err != nil {
		zap.L().Error("Failed to create creator history", zap.Error(err))
	}

	// Calculate ranks after adding the new creator
	if err := utils.CalculateCreatorRanks(h.db); err != nil {
		zap.L().Error("Failed to calculate ranks", zap.Error(err))
	}

	// Retrieve the creator again to get the calculated rank
	var creatorWithRank models.Creator
	if err := h.db.Where("id = ?", newCreator.ID).First(&creatorWithRank).Error; err != nil {
		zap.L().Error("Failed to retrieve creator with rank", zap.Error(err))
		// Return the creator without rank if retrieval fails
		return c.JSON(fiber.Map{
			"message": "Creator added successfully",
			"creator": newCreator,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Creator added successfully",
		"creator": creatorWithRank,
	})
}

func applyCreatorSearch(query *gorm.DB, search string) *gorm.DB {
	if search == "" {
		return query
	}

	return query.Where("username LIKE ? OR display_name LIKE ?", "%"+search+"%", "%"+search+"%")
}

func collectCreatorIDs(creators []models.Creator) []string {
	creatorIDs := make([]string, len(creators))
	for i, creator := range creators {
		creatorIDs[i] = creator.ID
	}

	return creatorIDs
}

func buildPagination(page, limit int, total int64) fiber.Map {
	return fiber.Map{
		"page":       page,
		"limit":      limit,
		"totalCount": total,
		"totalPages": (total + int64(limit) - 1) / int64(limit),
	}
}

func buildCreatorMetrics(creator models.Creator, snapshots map[string]models.CreatorHistory) creatorMetrics {
	metrics := creatorMetrics{
		MediaLikes: creator.MediaLikes,
		PostLikes:  creator.PostLikes,
		Followers:  creator.Followers,
		ImageCount: creator.ImageCount,
		VideoCount: creator.VideoCount,
	}

	if snapshot, ok := snapshots[creator.ID]; ok {
		metrics.MediaLikes = snapshot.MediaLikes
		metrics.PostLikes = snapshot.PostLikes
		metrics.Followers = snapshot.Followers
		metrics.ImageCount = snapshot.ImageCount
		metrics.VideoCount = snapshot.VideoCount
	}

	return metrics
}

func buildCreatorHistoryPoints(history []models.CreatorHistory) []CreatorHistoryPoint {
	historyPoints := make([]CreatorHistoryPoint, len(history))
	for i, point := range history {
		historyPoints[i] = CreatorHistoryPoint{
			ID:         point.ID,
			CreatorID:  point.CreatorID,
			MediaLikes: point.MediaLikes,
			PostLikes:  point.PostLikes,
			Followers:  point.Followers,
			ImageCount: point.ImageCount,
			VideoCount: point.VideoCount,
			CreatedAt:  point.CreatedAt.Unix(),
			UpdatedAt:  point.UpdatedAt.Unix(),
		}
	}

	return historyPoints
}

func buildCreatorWithHistory(
	creator models.Creator,
	snapshots map[string]models.CreatorHistory,
	history []models.CreatorHistory,
) CreatorWithHistory {
	response := buildCreatorSummary(creator, snapshots)
	response.History = buildCreatorHistoryPoints(history)
	return response
}

func buildCreatorsWithHistory(
	creators []models.Creator,
	snapshots map[string]models.CreatorHistory,
	historyByCreator map[string][]models.CreatorHistory,
) []CreatorWithHistory {
	creatorsWithHistory := make([]CreatorWithHistory, 0, len(creators))
	for _, creator := range creators {
		creatorsWithHistory = append(
			creatorsWithHistory,
			buildCreatorWithHistory(creator, snapshots, historyByCreator[creator.ID]),
		)
	}

	return creatorsWithHistory
}

func buildCreatorData(creators []models.Creator, snapshots map[string]models.CreatorHistory) []map[string]any {
	creatorsData := make([]map[string]any, len(creators))
	for i, creator := range creators {
		response := buildCreatorSummary(creator, snapshots)
		creatorsData[i] = map[string]any{
			"id":                response.ID,
			"username":          response.Username,
			"displayName":       response.DisplayName,
			"mediaLikes":        response.MediaLikes,
			"postLikes":         response.PostLikes,
			"followers":         response.Followers,
			"imageCount":        response.ImageCount,
			"videoCount":        response.VideoCount,
			"rank":              response.Rank,
			"lastCheckedAt":     response.LastCheckedAt,
			"isDeleted":         response.IsDeleted,
			"deletedDetectedAt": response.DeletedDetectedAt,
			"createdAt":         response.CreatedAt,
			"updatedAt":         response.UpdatedAt,
		}
	}

	return creatorsData
}

func buildCreatorSummary(
	creator models.Creator,
	snapshots map[string]models.CreatorHistory,
) CreatorWithHistory {
	metrics := buildCreatorMetrics(creator, snapshots)

	return CreatorWithHistory{
		ID:                creator.ID,
		Username:          creator.Username,
		DisplayName:       creator.DisplayName,
		MediaLikes:        metrics.MediaLikes,
		PostLikes:         metrics.PostLikes,
		Followers:         metrics.Followers,
		ImageCount:        metrics.ImageCount,
		VideoCount:        metrics.VideoCount,
		Rank:              creator.Rank,
		LastCheckedAt:     timeToUnixPtr(creator.LastCheckedAt),
		IsDeleted:         creator.IsDeleted,
		DeletedDetectedAt: timeToUnixPtr(creator.DeletedDetectedAt),
		CreatedAt:         creator.CreatedAt.Unix(),
		UpdatedAt:         creator.UpdatedAt.Unix(),
	}
}

func (h *CreatorHandler) loadCreatorSnapshotsForRange(
	creatorIDs []string,
	endDate *time.Time,
) (map[string]models.CreatorHistory, error) {
	if endDate == nil {
		return map[string]models.CreatorHistory{}, nil
	}

	return loadCreatorSnapshots(h.db, creatorIDs, *endDate)
}

func (h *CreatorHandler) loadCreatorHistoryByCreator(
	creatorIDs []string,
	startDate *time.Time,
	endDate *time.Time,
) (map[string][]models.CreatorHistory, error) {
	historyByCreator := make(map[string][]models.CreatorHistory)
	if len(creatorIDs) == 0 {
		return historyByCreator, nil
	}

	const batchSize = 1000
	for i := 0; i < len(creatorIDs); i += batchSize {
		end := min(i+batchSize, len(creatorIDs))
		batchIDs := creatorIDs[i:end]

		histQuery := h.db.Model(&models.CreatorHistory{}).
			Where("creator_id IN ?", batchIDs).
			Order("creator_id, created_at DESC")

		if startDate != nil {
			histQuery = histQuery.Where("created_at >= ?", *startDate)
		}
		if endDate != nil {
			histQuery = histQuery.Where("created_at <= ?", *endDate)
		}

		var batchHistories []models.CreatorHistory
		if err := histQuery.Find(&batchHistories).Error; err != nil {
			return nil, err
		}

		for _, history := range batchHistories {
			historyByCreator[history.CreatorID] = append(historyByCreator[history.CreatorID], history)
		}
	}

	return historyByCreator, nil
}

func loadCreatorSnapshots(db *gorm.DB, creatorIDs []string, endDate time.Time) (map[string]models.CreatorHistory, error) {
	if len(creatorIDs) == 0 {
		return map[string]models.CreatorHistory{}, nil
	}

	var snapshots []models.CreatorHistory
	if err := db.Table("creators AS c").
		Select("ch.id, ch.creator_id, ch.media_likes, ch.post_likes, ch.followers, ch.image_count, ch.video_count, ch.created_at, ch.updated_at").
		Joins(latestCreatorSnapshotJoinClause(), endDate).
		Where("c.id IN ?", creatorIDs).
		Where("ch.id IS NOT NULL").
		Scan(&snapshots).Error; err != nil {
		return nil, err
	}

	snapshotByCreator := make(map[string]models.CreatorHistory, len(snapshots))
	for _, snapshot := range snapshots {
		snapshotByCreator[snapshot.CreatorID] = snapshot
	}

	return snapshotByCreator, nil
}

func latestCreatorSnapshotJoinClause() string {
	return "LEFT JOIN creator_history AS ch ON ch.id = (" +
		"SELECT snapshot.id FROM creator_history AS snapshot FORCE INDEX (idx_creator_history_creator_created_id) " +
		"WHERE snapshot.creator_id = c.id AND snapshot.created_at <= ? " +
		"ORDER BY snapshot.created_at DESC, snapshot.id DESC LIMIT 1)"
}
