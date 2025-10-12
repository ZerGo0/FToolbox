# GORM Model Pattern

## When to Use
Use this pattern for all database entities to ensure consistent schema management, proper indexing, and automatic migrations.

## Why It Exists
This pattern leverages GORM's AutoMigrate feature for schema evolution while maintaining consistent naming conventions, proper indexing, and JSON serialization.

## Implementation
Models are Go structs with GORM tags that define database schema. Key characteristics:

- Primary keys as string IDs with varchar(255) type
- Consistent column naming (snake_case in database, camelCase in Go)
- Proper indexing for frequently queried columns
- Timestamp fields with automatic defaults
- JSON tags for API serialization
- Custom TableName() methods for explicit table names
- Pointer types for optional fields

## References
- `backend-go/models/creator.go:7-26` - Creator model with comprehensive field mapping
- `backend-go/models/tag.go:7-25` - Tag model with complex indexing strategy
- `backend-go/models/worker.go` - Worker model for status tracking
- `backend-go/database/migrate.go:9-20` - AutoMigrate usage for all models
- `backend-go/database/connection.go:12-30` - Database connection with GORM configuration

## Key Conventions
- Use `gorm:"primaryKey"` for primary keys
- Apply `gorm:"index"` for searchable columns
- Use `gorm:"not null;default:false"` for boolean fields
- Implement `TableName()` method for explicit table naming
- Use pointer types (*string, *int, *time.Time) for nullable fields
- Exclude internal fields from JSON with `json:"-"`
- Use composite indexes for multi-column queries