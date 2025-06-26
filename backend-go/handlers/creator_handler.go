package handlers

import (
	"ftoolbox/fansly"
	"ftoolbox/models"
	"ftoolbox/utils"
	"strconv"
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

func (h *CreatorHandler) GetCreators(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	sortBy := c.Query("sortBy", "followers")
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

	var creators []models.Creator
	query := h.db.Model(&models.Creator{})

	if search != "" {
		query = query.Where("username LIKE ? OR display_name LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// When sorting by rank, exclude creators without a rank from the base query
	if sortBy == "rank" {
		query = query.Where("rank IS NOT NULL")
	}

	var total int64
	query.Count(&total)

	// Handle sorting
	needsHistory := includeHistory

	// Map frontend sortBy values to database columns
	columnMap := map[string]string{
		"followers":  "followers",
		"mediaLikes": "media_likes",
		"postLikes":  "post_likes",
		"imageCount": "image_count",
		"videoCount": "video_count",
		"updatedAt":  "updated_at",
		"username":   "username",
		"rank":       "rank",
	}

	// Handle sorting
	dbColumn, ok := columnMap[sortBy]
	if !ok {
		dbColumn = "followers" // default
	}

	orderClause := dbColumn + " " + sortOrder
	// Special case: when sorting by followers desc, also sort by created_at desc as secondary
	if sortBy == "followers" && sortOrder == "desc" {
		orderClause = "followers DESC, created_at DESC"
	}
	query = query.Order(orderClause).Limit(limit).Offset(offset)

	if err := query.Find(&creators).Error; err != nil {
		zap.L().Error("Failed to fetch creators", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch creators"})
	}

	// If we need history, fetch it for each creator
	var creatorsWithHistory []CreatorWithHistory
	if needsHistory {
		// Collect all creator IDs for batch loading
		creatorIDs := make([]string, len(creators))
		creatorMap := make(map[string]models.Creator)
		for i, creator := range creators {
			creatorIDs[i] = creator.ID
			creatorMap[creator.ID] = creator
		}

		// Build base query for all histories
		histQuery := h.db.Model(&models.CreatorHistory{}).
			Where("creator_id IN ?", creatorIDs).
			Order("creator_id, created_at DESC")

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

		// Fetch all histories in one query
		var allHistories []models.CreatorHistory
		if err := histQuery.Find(&allHistories).Error; err != nil {
			zap.L().Error("Failed to fetch creator histories", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch creator histories"})
		}

		// Group histories by creator ID
		historyByCreator := make(map[string][]models.CreatorHistory)
		for _, hist := range allHistories {
			historyByCreator[hist.CreatorID] = append(historyByCreator[hist.CreatorID], hist)
		}

		// Process each creator with its history
		for _, creator := range creators {
			creatorWithHist := CreatorWithHistory{
				ID:                creator.ID,
				Username:          creator.Username,
				DisplayName:       creator.DisplayName,
				MediaLikes:        creator.MediaLikes,
				PostLikes:         creator.PostLikes,
				Followers:         creator.Followers,
				ImageCount:        creator.ImageCount,
				VideoCount:        creator.VideoCount,
				Rank:              creator.Rank,
				LastCheckedAt:     timeToUnixPtr(creator.LastCheckedAt),
				IsDeleted:         creator.IsDeleted,
				DeletedDetectedAt: timeToUnixPtr(creator.DeletedDetectedAt),
				CreatedAt:         creator.CreatedAt.Unix(),
				UpdatedAt:         creator.UpdatedAt.Unix(),
			}

			history := historyByCreator[creator.ID]

			// Convert to CreatorHistoryPoint
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

			if includeHistory {
				creatorWithHist.History = historyPoints
			}

			creatorsWithHistory = append(creatorsWithHistory, creatorWithHist)
		}
	}

	// Return response with consistent format
	if needsHistory {
		return c.JSON(fiber.Map{
			"creators": creatorsWithHistory,
			"pagination": fiber.Map{
				"page":       page,
				"limit":      limit,
				"totalCount": total,
				"totalPages": (total + int64(limit) - 1) / int64(limit),
			},
		})
	}

	// For non-history responses, we need to convert creators to proper format
	creatorsData := make([]map[string]interface{}, len(creators))
	for i, creator := range creators {
		creatorsData[i] = map[string]interface{}{
			"id":                creator.ID,
			"username":          creator.Username,
			"displayName":       creator.DisplayName,
			"mediaLikes":        creator.MediaLikes,
			"postLikes":         creator.PostLikes,
			"followers":         creator.Followers,
			"imageCount":        creator.ImageCount,
			"videoCount":        creator.VideoCount,
			"rank":              creator.Rank,
			"lastCheckedAt":     timeToUnixPtr(creator.LastCheckedAt),
			"isDeleted":         creator.IsDeleted,
			"deletedDetectedAt": timeToUnixPtr(creator.DeletedDetectedAt),
			"createdAt":         creator.CreatedAt.Unix(),
			"updatedAt":         creator.UpdatedAt.Unix(),
		}
	}

	return c.JSON(fiber.Map{
		"creators": creatorsData,
		"pagination": fiber.Map{
			"page":       page,
			"limit":      limit,
			"totalCount": total,
			"totalPages": (total + int64(limit) - 1) / int64(limit),
		},
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
