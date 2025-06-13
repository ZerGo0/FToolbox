package utils

import "gorm.io/gorm"

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
