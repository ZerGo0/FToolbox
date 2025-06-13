package main

import (
	"database/sql"
	"fmt"
	"ftoolbox/config"
	"ftoolbox/database"
	"ftoolbox/models"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run migrate.go <path-to-sqlite-db>")
	}

	sqlitePath := os.Args[1]

	// Load config
	cfg := config.Load()

	// Drop existing database if it exists
	fmt.Println("Dropping existing database if it exists...")
	dropDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort))
	if err != nil {
		log.Fatal("Failed to connect to MariaDB server:", err)
	}
	defer dropDB.Close()

	_, err = dropDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", cfg.DBDatabase))
	if err != nil {
		log.Fatal("Failed to drop database:", err)
	}

	_, err = dropDB.Exec(fmt.Sprintf("CREATE DATABASE `%s`", cfg.DBDatabase))
	if err != nil {
		log.Fatal("Failed to create database:", err)
	}
	fmt.Printf("Created fresh database: %s\n", cfg.DBDatabase)

	// Now connect to the new database
	mariaDB, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to MariaDB:", err)
	}

	// Run migrations
	if err := database.AutoMigrate(mariaDB); err != nil {
		log.Fatal("Failed to run MariaDB migrations:", err)
	}

	// Connect to SQLite
	sqliteDB, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		log.Fatal("Failed to connect to SQLite:", err)
	}
	defer sqliteDB.Close()

	// Migrate data
	fmt.Println("Starting migration from SQLite to MariaDB...")

	// Migrate tags
	if err := migrateTags(sqliteDB, mariaDB); err != nil {
		log.Fatal("Failed to migrate tags:", err)
	}

	// Migrate tag history
	if err := migrateTagHistory(sqliteDB, mariaDB); err != nil {
		log.Fatal("Failed to migrate tag history:", err)
	}

	// Migrate workers
	if err := migrateWorkers(sqliteDB, mariaDB); err != nil {
		log.Fatal("Failed to migrate workers:", err)
	}

	fmt.Println("Migration completed successfully!")
}

func migrateTags(sqliteDB *sql.DB, mariaDB *gorm.DB) error {
	fmt.Println("Migrating tags...")

	rows, err := sqliteDB.Query(`
		SELECT id, tag, view_count, rank, fansly_created_at, last_checked_at, 
		       last_used_for_discovery, created_at, updated_at
		FROM tags
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var tag models.Tag
		var id string
		var tagName string
		var viewCount int64
		var rank sql.NullInt32
		var fanslyCreatedAt int64
		var lastCheckedAt, lastUsedForDiscovery, createdAt, updatedAt sql.NullInt64

		err := rows.Scan(&id, &tagName, &viewCount, &rank, &fanslyCreatedAt,
			&lastCheckedAt, &lastUsedForDiscovery, &createdAt, &updatedAt)
		if err != nil {
			return err
		}

		tag.ID = id
		tag.Tag = tagName
		tag.ViewCount = viewCount

		if rank.Valid {
			r := int(rank.Int32)
			tag.Rank = &r
		}

		// Convert Unix timestamps to time.Time
		tag.FanslyCreatedAt = time.Unix(fanslyCreatedAt, 0)

		if lastCheckedAt.Valid {
			t := time.Unix(lastCheckedAt.Int64, 0)
			tag.LastCheckedAt = &t
		}

		if lastUsedForDiscovery.Valid {
			t := time.Unix(lastUsedForDiscovery.Int64, 0)
			tag.LastUsedForDiscovery = &t
		}

		if createdAt.Valid {
			tag.CreatedAt = time.Unix(createdAt.Int64, 0)
		} else {
			tag.CreatedAt = time.Now()
		}

		if updatedAt.Valid {
			tag.UpdatedAt = time.Unix(updatedAt.Int64, 0)
		} else {
			tag.UpdatedAt = time.Now()
		}

		if err := mariaDB.Create(&tag).Error; err != nil {
			log.Printf("Error migrating tag %s: %v", tagName, err)
			continue
		}

		count++
	}

	fmt.Printf("Migrated %d tags\n", count)
	return nil
}

func migrateTagHistory(sqliteDB *sql.DB, mariaDB *gorm.DB) error {
	fmt.Println("Migrating tag history...")

	rows, err := sqliteDB.Query(`
		SELECT id, tag_id, view_count, change, created_at, updated_at
		FROM tag_history
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var history models.TagHistory
		var id uint
		var tagID string
		var viewCount, change int64
		var createdAt, updatedAt sql.NullInt64

		err := rows.Scan(&id, &tagID, &viewCount, &change, &createdAt, &updatedAt)
		if err != nil {
			return err
		}

		history.ID = id
		history.TagID = tagID
		history.ViewCount = viewCount
		history.Change = change

		if createdAt.Valid {
			history.CreatedAt = time.Unix(createdAt.Int64, 0)
		} else {
			history.CreatedAt = time.Now()
		}

		if updatedAt.Valid {
			history.UpdatedAt = time.Unix(updatedAt.Int64, 0)
		} else {
			history.UpdatedAt = time.Now()
		}

		if err := mariaDB.Create(&history).Error; err != nil {
			log.Printf("Error migrating tag history %d: %v", id, err)
			continue
		}

		count++
	}

	fmt.Printf("Migrated %d tag history records\n", count)
	return nil
}

func migrateWorkers(sqliteDB *sql.DB, mariaDB *gorm.DB) error {
	fmt.Println("Migrating workers...")

	rows, err := sqliteDB.Query(`
		SELECT id, name, last_run_at, next_run_at, status, last_error, 
		       run_count, success_count, failure_count, is_enabled, created_at, updated_at
		FROM workers
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var worker models.Worker
		var id uint
		var name, status string
		var lastRunAt, nextRunAt, createdAt, updatedAt sql.NullInt64
		var lastError sql.NullString
		var runCount, successCount, failureCount int
		var isEnabled int

		err := rows.Scan(&id, &name, &lastRunAt, &nextRunAt, &status, &lastError,
			&runCount, &successCount, &failureCount, &isEnabled, &createdAt, &updatedAt)
		if err != nil {
			return err
		}

		worker.ID = id
		worker.Name = name
		worker.Status = status
		worker.RunCount = runCount
		worker.SuccessCount = successCount
		worker.FailureCount = failureCount
		worker.IsEnabled = isEnabled != 0

		if lastRunAt.Valid {
			t := time.Unix(lastRunAt.Int64, 0)
			worker.LastRunAt = &t
		}

		if nextRunAt.Valid {
			t := time.Unix(nextRunAt.Int64, 0)
			worker.NextRunAt = &t
		}

		if lastError.Valid {
			worker.LastError = &lastError.String
		}

		if createdAt.Valid {
			worker.CreatedAt = time.Unix(createdAt.Int64, 0)
		} else {
			worker.CreatedAt = time.Now()
		}

		if updatedAt.Valid {
			worker.UpdatedAt = time.Unix(updatedAt.Int64, 0)
		} else {
			worker.UpdatedAt = time.Now()
		}

		if err := mariaDB.Create(&worker).Error; err != nil {
			log.Printf("Error migrating worker %s: %v", name, err)
			continue
		}

		count++
	}

	fmt.Printf("Migrated %d workers\n", count)
	return nil
}
