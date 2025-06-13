package models

import (
	"time"
)

type Worker struct {
	ID           uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name         string     `gorm:"unique;not null;column:name" json:"name"`
	LastRunAt    *time.Time `gorm:"column:last_run_at" json:"lastRunAt,omitempty"`
	NextRunAt    *time.Time `gorm:"column:next_run_at" json:"nextRunAt,omitempty"`
	Status       string     `gorm:"not null;default:'idle';column:status" json:"status"` // idle, running, failed
	LastError    *string    `gorm:"column:last_error" json:"lastError,omitempty"`
	RunCount     int        `gorm:"not null;default:0;column:run_count" json:"runCount"`
	SuccessCount int        `gorm:"not null;default:0;column:success_count" json:"successCount"`
	FailureCount int        `gorm:"not null;default:0;column:failure_count" json:"failureCount"`
	IsEnabled    bool       `gorm:"not null;default:true;column:is_enabled" json:"isEnabled"`
	CreatedAt    time.Time  `gorm:"not null;column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"not null;column:updated_at;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (Worker) TableName() string {
	return "workers"
}
