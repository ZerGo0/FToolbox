# Database Status Tracking Pattern

## When to Use
Use this pattern for atomically updating worker status counters and statistics in the database to ensure thread-safe increments without race conditions.

## Why It Exists
Provides atomic database operations for counters that prevent race conditions when multiple goroutines update the same statistics simultaneously. Ensures data integrity in concurrent worker environments.

## Implementation
Use GORM's `gorm.Expr` for atomic counter updates:

```go
// Increment failure count
result := db.Model(&Worker{}).Where("id = ?", workerID).Update("failure_count", gorm.Expr("failure_count + 1"))

// Increment success count  
result := db.Model(&Worker{}).Where("id = ?", workerID).Update("success_count", gorm.Expr("success_count + 1"))

// Update last run timestamp
result := db.Model(&Worker{}).Where("id = ?", workerID).Update("last_run_at", time.Now())
```

## Source References
- `backend-go/workers/manager.go:166` - Failure count increment in worker execution
- `backend-go/workers/manager.go:269` - Success count increment after worker completion
- `backend-go/workers/manager.go:277` - Status update with timestamp

## Key Conventions
- Always use `gorm.Expr("column_name + 1")` for atomic increments
- Combine counter updates with timestamp updates for complete status tracking
- Check `result.Error` for database operation failures
- Use specific WHERE clauses to target exact records
- Update both counters and timestamps in the same operation when possible
- Log errors immediately if atomic updates fail