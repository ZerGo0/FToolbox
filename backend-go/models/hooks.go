package models

import (
	"gorm.io/gorm"
)

// CalculateTagRanks recalculates ranks for all tags
func CalculateTagRanks(db *gorm.DB) error {
	// Use raw SQL for better performance and to avoid hooks
	// DENSE_RANK() ensures no gaps in ranking when there are ties
	sql := `
		UPDATE tags t1
		JOIN (
			SELECT 
				id,
				DENSE_RANK() OVER (ORDER BY view_count DESC, created_at ASC) as new_rank
			FROM tags
		) t2 ON t1.id = t2.id
		SET t1.rank = t2.new_rank
	`

	return db.Exec(sql).Error
}

// AfterCreate is called after creating a tag
func (t *Tag) AfterCreate(tx *gorm.DB) error {
	// Calculate rank for the new tag
	var rank int
	err := tx.Raw(`
		SELECT COUNT(*) + 1 
		FROM tags 
		WHERE view_count > ? OR (view_count = ? AND created_at < ?)
	`, t.ViewCount, t.ViewCount, t.CreatedAt).Scan(&rank).Error

	if err != nil {
		return err
	}

	// Update only this tag's rank to avoid recursion
	return tx.Exec("UPDATE tags SET rank = ? WHERE id = ?", rank, t.ID).Error
}

// AfterUpdate is called after updating a tag
func (t *Tag) AfterUpdate(tx *gorm.DB) error {
	// Check if view_count was changed
	var oldViewCount int64
	tx.Raw("SELECT view_count FROM tags WHERE id = ?", t.ID).Scan(&oldViewCount)

	// Only recalculate all ranks if view count changed
	if oldViewCount != t.ViewCount {
		return CalculateTagRanks(tx)
	}
	return nil
}
