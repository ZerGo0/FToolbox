package models

import "time"

// TagRelationDaily stores daily co-usage counts of tags observed during discovery
type TagRelationDaily struct {
    TagID        string    `gorm:"primaryKey;type:varchar(255);column:tag_id;index:idx_trd_tag_bucket,priority:1" json:"tagId"`
    RelatedTagID string    `gorm:"primaryKey;type:varchar(255);column:related_tag_id" json:"relatedTagId"`
    BucketDate   time.Time `gorm:"primaryKey;type:date;column:bucket_date;index:idx_tag_relations_daily_bucket;index:idx_trd_tag_bucket,priority:2" json:"bucketDate"`
    CoCount      int64     `gorm:"not null;default:0;column:co_count" json:"coCount"`
    LastSeenAt   time.Time `gorm:"not null;column:last_seen_at;default:CURRENT_TIMESTAMP" json:"lastSeenAt"`

    // Helpful indexes
    // index on bucket for quick purge
    // gorm indexes via struct tags are attached to fields; define here for clarity
    // BucketDate has implicit index defined below
}

func (TagRelationDaily) TableName() string {
    return "tag_relations_daily"
}
