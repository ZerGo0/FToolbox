# GORM Model Conventions Pattern

## When to Use
Use this pattern for all database models to ensure consistent field naming, indexing, and JSON serialization.

## Why It Exists
This pattern standardizes database schema conventions, ensures proper indexing for performance, and maintains consistent API response formats.

## Implementation Details

### Model Structure
```go
type Model struct {
    ID                string     `gorm:"primaryKey;type:varchar(255);column:id" json:"id"`
    // Business fields
    FieldName         string     `gorm:"not null;index;column:field_name" json:"fieldName"`
    OptionalField     *string    `gorm:"column:optional_field" json:"optionalField"`
    // Timestamps
    CreatedAt         time.Time  `gorm:"not null;column:created_at;default:CURRENT_TIMESTAMP" json:"-"`
    UpdatedAt         time.Time  `gorm:"not null;column:updated_at;default:CURRENT_TIMESTAMP" json:"-"`
}
```

### Field Naming Conventions
- Database columns: `snake_case`
- JSON fields: `camelCase`
- Go fields: `PascalCase`

### Indexing Strategy
```go
// Single field indexes
ViewCount    int64 `gorm:"not null;column:view_count;index" json:"viewCount"`

// Composite indexes for common query patterns
CreatedAt    time.Time `gorm:"not null;column:created_at;index:idx_model_view_created,priority:2" json:"-"`
ViewCount    int64    `gorm:"not null;column:view_count;index:idx_model_view_created,priority:1" json:"viewCount"`
```

### Timestamp Handling
Hide internal timestamps from JSON responses, convert to Unix timestamps manually:
```go
// In handler responses
CreatedAt:    model.CreatedAt.Unix(),
UpdatedAt:    model.UpdatedAt.Unix(),
LastCheckedAt: timeToUnixPtr(model.LastCheckedAt),
```

### Soft Deletes Pattern
```go
IsDeleted         bool       `gorm:"not null;default:false;column:is_deleted;index:idx_model_is_deleted_deleted,priority:1" json:"isDeleted"`
DeletedDetectedAt *time.Time `gorm:"column:deleted_detected_at;index:idx_model_is_deleted_deleted,priority:2" json:"deletedDetectedAt"`
```

### Table Name Specification
Always specify table names explicitly:
```go
func (Model) TableName() string {
    return "table_name"
}
```

## References
- `backend-go/models/tag.go:7-26` - Tag model with indexing strategy
- `backend-go/models/creator.go:7-27` - Creator model conventions
- `backend-go/handlers/tag_handler.go:808-821` - Timestamp conversion utilities