package models

import (
	"time"
)

type TagHistory struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	TagID     string    `gorm:"not null;index;type:varchar(255);column:tag_id" json:"tagId"`
	ViewCount int64     `gorm:"not null;column:view_count" json:"viewCount"`
	Change    int64     `gorm:"not null;column:change" json:"change"`
	PostCount int64     `gorm:"not null;default:0;column:post_count" json:"postCount"`
	CreatedAt time.Time `gorm:"not null;column:created_at;default:CURRENT_TIMESTAMP;index" json:"-"`
	UpdatedAt time.Time `gorm:"not null;column:updated_at;default:CURRENT_TIMESTAMP" json:"-"`
}

func (TagHistory) TableName() string {
	return "tag_history"
}
