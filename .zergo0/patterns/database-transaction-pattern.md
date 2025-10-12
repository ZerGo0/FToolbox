# Database Transaction Pattern

## When to Use
Use this pattern for any database operations that need to update multiple tables atomically to ensure data consistency and prevent partial updates.

## Why It Exists
Database transactions ensure that multiple related operations either all succeed or all fail together, maintaining data integrity and preventing inconsistent states in the database.

## Implementation

### Transaction Structure
Begin transaction, perform operations, commit or rollback:
```go
tx := w.db.Begin()

// Perform database operations
if err := tx.Save(&model).Error; err != nil {
    tx.Rollback()
    return fmt.Errorf("failed to update model: %w", err)
}

if err := tx.Create(&history).Error; err != nil {
    tx.Rollback()
    return fmt.Errorf("failed to create history: %w", err)
}

// Commit transaction
if err := tx.Commit().Error; err != nil {
    return fmt.Errorf("failed to commit transaction: %w", err)
}
```

### Error Handling Pattern
Always rollback on error and wrap errors with context:
```go
if err := tx.Operation(&data).Error; err != nil {
    tx.Rollback()
    return fmt.Errorf("failed to perform operation: %w", err)
}
```

### Common Use Cases
- Creating history records alongside model updates
- Updating related entities in multiple tables
- Ensuring consistency between main data and statistics

## Source References
- `backend-go/workers/tag_updater.go:123-153` - Tag update with history creation
- `backend-go/workers/creator_updater.go:120-162` - Creator creation with history
- `backend-go/workers/creator_updater.go:183-223` - Creator update with history
- `backend-go/workers/statistics_calculator.go:52-112` - Statistics calculation transaction
- `backend-go/workers/statistics_calculator.go:129-200` - Creator statistics transaction

## Key Conventions
- Always begin with `tx := w.db.Begin()`
- Immediately rollback on any error: `tx.Rollback()`
- Use descriptive error messages with `fmt.Errorf` and `%w` for error wrapping
- Commit only after all operations succeed: `tx.Commit()`
- Keep transactions as short as possible to avoid locking
- Use transactions for any multi-table operations that must be atomic
- Log transaction success/failure appropriately for debugging