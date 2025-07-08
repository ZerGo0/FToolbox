package models

import (
	"time"
)

type CreatorStatistics struct {
	ID                         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TotalFollowers             int64     `gorm:"not null;column:total_followers" json:"totalFollowers"`
	FollowersChange24h         int64     `gorm:"not null;default:0;column:followers_change_24h" json:"followersChange24h"`
	FollowersChangePercent24h  float64   `gorm:"not null;default:0;column:followers_change_percent_24h" json:"followersChangePercent24h"`
	TotalMediaLikes            int64     `gorm:"not null;column:total_media_likes" json:"totalMediaLikes"`
	MediaLikesChange24h        int64     `gorm:"not null;default:0;column:media_likes_change_24h" json:"mediaLikesChange24h"`
	MediaLikesChangePercent24h float64   `gorm:"not null;default:0;column:media_likes_change_percent_24h" json:"mediaLikesChangePercent24h"`
	TotalPostLikes             int64     `gorm:"not null;column:total_post_likes" json:"totalPostLikes"`
	PostLikesChange24h         int64     `gorm:"not null;default:0;column:post_likes_change_24h" json:"postLikesChange24h"`
	PostLikesChangePercent24h  float64   `gorm:"not null;default:0;column:post_likes_change_percent_24h" json:"postLikesChangePercent24h"`
	CalculatedAt               time.Time `gorm:"not null;column:calculated_at" json:"calculatedAt"`
	CreatedAt                  time.Time `gorm:"not null;column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt                  time.Time `gorm:"not null;column:updated_at;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (CreatorStatistics) TableName() string {
	return "creator_statistics"
}