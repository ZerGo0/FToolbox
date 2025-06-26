package models

import (
	"time"
)

type Creator struct {
	ID                string     `gorm:"primaryKey;type:varchar(255);column:id" json:"id"`
	Username          string     `gorm:"not null;index;column:username" json:"username"`
	DisplayName       *string    `gorm:"column:display_name" json:"displayName"`
	MediaLikes        int64      `gorm:"not null;column:media_likes" json:"mediaLikes"`
	PostLikes         int64      `gorm:"not null;column:post_likes" json:"postLikes"`
	Followers         int64      `gorm:"not null;column:followers" json:"followers"`
	ImageCount        int64      `gorm:"not null;column:image_count" json:"imageCount"`
	VideoCount        int64      `gorm:"not null;column:video_count" json:"videoCount"`
	Rank              *int       `gorm:"column:rank;index" json:"rank"`
	LastCheckedAt     *time.Time `gorm:"column:last_checked_at" json:"-"`
	IsDeleted         bool       `gorm:"not null;default:false;column:is_deleted" json:"isDeleted"`
	DeletedDetectedAt *time.Time `gorm:"column:deleted_detected_at" json:"deletedDetectedAt"`
	CreatedAt         time.Time  `gorm:"not null;column:created_at;default:CURRENT_TIMESTAMP" json:"-"`
	UpdatedAt         time.Time  `gorm:"not null;column:updated_at;default:CURRENT_TIMESTAMP" json:"-"`
}

func (Creator) TableName() string {
	return "creators"
}
