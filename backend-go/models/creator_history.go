package models

import (
	"time"
)

type CreatorHistory struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatorID  string    `gorm:"not null;index;type:varchar(255);column:creator_id" json:"creatorId"`
	MediaLikes int64     `gorm:"not null;column:media_likes" json:"mediaLikes"`
	PostLikes  int64     `gorm:"not null;column:post_likes" json:"postLikes"`
	Followers  int64     `gorm:"not null;column:followers" json:"followers"`
	ImageCount int64     `gorm:"not null;column:image_count" json:"imageCount"`
	VideoCount int64     `gorm:"not null;column:video_count" json:"videoCount"`
	CreatedAt  time.Time `gorm:"not null;column:created_at;default:CURRENT_TIMESTAMP;index" json:"-"`
	UpdatedAt  time.Time `gorm:"not null;column:updated_at;default:CURRENT_TIMESTAMP" json:"-"`
}

func (CreatorHistory) TableName() string {
	return "creator_history"
}
