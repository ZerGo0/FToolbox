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

// CalculateCreatorRanks recalculates ranks for all creators
func CalculateCreatorRanks(db *gorm.DB) error {
	// Use raw SQL for better performance and to avoid hooks
	// DENSE_RANK() ensures no gaps in ranking when there are ties
	// Rank by followers as the primary metric, with created_at as tiebreaker
	sql := `
		UPDATE creators c1
		JOIN (
			SELECT 
				id,
				DENSE_RANK() OVER (ORDER BY followers DESC, created_at ASC) as new_rank
			FROM creators
			WHERE is_deleted = 0
		) c2 ON c1.id = c2.id
		SET c1.rank = c2.new_rank
	`

	// Also clear ranks for deleted creators
	if err := db.Exec(sql).Error; err != nil {
		return err
	}

	// Clear ranks for deleted creators
	clearSql := `UPDATE creators SET rank = NULL WHERE is_deleted = 1`
	return db.Exec(clearSql).Error
}
