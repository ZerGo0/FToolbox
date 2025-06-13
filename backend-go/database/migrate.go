package database

import (
	"ftoolbox/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Tag{},
		&models.TagHistory{},
		&models.Worker{},
	)
}
