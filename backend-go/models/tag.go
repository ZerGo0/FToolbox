package models

import (
	"time"
)

type Tag struct {
	ID                   string     `gorm:"primaryKey;type:varchar(255);column:id" json:"id"`
	Tag                  string     `gorm:"unique;not null;index;column:tag" json:"tag"`
	ViewCount            int64      `gorm:"not null;column:view_count" json:"viewCount"`
	Rank                 *int       `gorm:"column:rank;index" json:"rank"`
	FanslyCreatedAt      time.Time  `gorm:"not null;column:fansly_created_at" json:"-"`
	LastCheckedAt        *time.Time `gorm:"column:last_checked_at" json:"-"`
	LastUsedForDiscovery *time.Time `gorm:"column:last_used_for_discovery" json:"-"`
	CreatedAt            time.Time  `gorm:"not null;column:created_at;default:CURRENT_TIMESTAMP" json:"-"`
	UpdatedAt            time.Time  `gorm:"not null;column:updated_at;default:CURRENT_TIMESTAMP" json:"-"`
}

func (Tag) TableName() string {
	return "tags"
}
