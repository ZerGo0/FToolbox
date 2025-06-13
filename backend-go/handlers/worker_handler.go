package handlers

import (
	"ftoolbox/models"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WorkerHandler struct {
	db *gorm.DB
}

func NewWorkerHandler(db *gorm.DB) *WorkerHandler {
	return &WorkerHandler{db: db}
}

func (h *WorkerHandler) GetStatus(c *fiber.Ctx) error {
	var workers []models.Worker

	if err := h.db.Find(&workers).Error; err != nil {
		zap.L().Error("Failed to fetch workers", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch worker status"})
	}

	// Check if any worker is running
	isRunning := false
	hasFailed := false
	for _, w := range workers {
		if w.Status == "running" {
			isRunning = true
		}
		if w.Status == "failed" {
			hasFailed = true
		}
	}

	// Determine overall status
	status := "idle"
	if hasFailed {
		status = "failed"
	} else if isRunning {
		status = "running"
	}

	return c.JSON(fiber.Map{
		"status": status,
	})
}
