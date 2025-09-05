package models

import (
	"time"
)

type Tag struct {
    ID                   string     `gorm:"primaryKey;type:varchar(255);column:id" json:"id"`
    Tag                  string     `gorm:"unique;not null;index;column:tag" json:"tag"`
    ViewCount            int64      `gorm:"not null;column:view_count;index:idx_tags_view_created,priority:1" json:"viewCount"`
    PostCount            int64      `gorm:"not null;default:0;column:post_count;index:idx_tags_post_created,priority:1" json:"postCount"`
    Rank                 *int       `gorm:"column:rank;index" json:"rank"`
    Heat                 float64    `gorm:"not null;default:0;column:heat;index" json:"heat"`
    Debug                bool       `gorm:"not null;default:false;column:debug" json:"debug"`
    FanslyCreatedAt      time.Time  `gorm:"not null;column:fansly_created_at" json:"-"`
    LastCheckedAt        *time.Time `gorm:"column:last_checked_at" json:"-"`
    LastUsedForDiscovery *time.Time `gorm:"column:last_used_for_discovery" json:"-"`
    IsDeleted            bool       `gorm:"not null;default:false;column:is_deleted;index:idx_tags_is_deleted_deleted,priority:1" json:"isDeleted"`
    DeletedDetectedAt    *time.Time `gorm:"column:deleted_detected_at;index:idx_tags_is_deleted_deleted,priority:2" json:"deletedDetectedAt"`
    CreatedAt            time.Time  `gorm:"not null;column:created_at;default:CURRENT_TIMESTAMP;index:idx_tags_view_created,priority:2;index:idx_tags_post_created,priority:2" json:"-"`
    UpdatedAt            time.Time  `gorm:"not null;column:updated_at;default:CURRENT_TIMESTAMP;index:idx_tags_updated_at" json:"-"`
}

func (Tag) TableName() string {
	return "tags"
}
