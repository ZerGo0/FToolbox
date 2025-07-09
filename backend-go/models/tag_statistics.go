package models

import (
	"time"
)

type TagStatistics struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TotalViewCount   int64     `gorm:"not null;column:total_view_count" json:"totalViewCount"`
	Change24h        int64     `gorm:"not null;default:0;column:change_24h" json:"change24h"`
	ChangePercent24h float64   `gorm:"not null;default:0;column:change_percent_24h" json:"changePercent24h"`
	CalculatedAt     time.Time `gorm:"not null;column:calculated_at" json:"calculatedAt"`
	CreatedAt        time.Time `gorm:"not null;column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"not null;column:updated_at;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (TagStatistics) TableName() string {
	return "tag_statistics"
}
