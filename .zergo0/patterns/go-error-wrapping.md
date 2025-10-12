# Go Error Wrapping Pattern

## When to Use
Use this pattern for all error handling in Go code to provide context, maintain error chains, and enable proper error debugging and logging.

## Why It Exists
Error wrapping with `fmt.Errorf` and `%w` preserves the original error while adding context about where and why the error occurred. This enables better debugging, error handling, and user feedback.

## Implementation

### Basic Error Wrapping
Wrap errors with context using `%w` verb:
```go
if err := operation(); err != nil {
    return fmt.Errorf("failed to perform operation: %w", err)
}
```

### Database Operation Errors
Wrap database errors with specific context:
```go
if err := db.Save(&model).Error; err != nil {
    return fmt.Errorf("failed to update tag: %w", err)
}

if err := tx.Create(&history).Error; err != nil {
    return fmt.Errorf("failed to create history: %w", err)
}

if err := tx.Commit().Error; err != nil {
    return fmt.Errorf("failed to commit transaction: %w", err)
}
```

### API Client Errors
Wrap external API errors with request context:
```go
if err := c.globalLimiter.Wait(ctx); err != nil {
    return nil, fmt.Errorf("rate limiter error: %w", err)
}

if err := json.Unmarshal(body, &response); err != nil {
    return nil, fmt.Errorf("failed to decode response: %w", err)
}
```

### Worker Management Errors
Wrap worker operation errors with worker context:
```go
if err := m.db.Create(&dbWorker).Error; err != nil {
    return fmt.Errorf("failed to create worker record: %w", err)
}

if err := m.db.Where("name = ?", name).First(&dbWorker).Error; err != nil {
    return fmt.Errorf("worker %s not found", name)
}
```

### Connection Errors
Wrap connection errors with service context:
```go
if err != nil {
    return nil, fmt.Errorf("failed to connect to database: %w", err)
}
```

## Source References
- `backend-go/workers/tag_updater.go:43,102,107,134,147,152` - Worker operation error wrapping
- `backend-go/workers/manager.go:42,58,61,81,87,93,118` - Worker management errors
- `backend-go/fansly/client.go:66,72,113,119,143,149,153` - API client error wrapping
- `backend-go/database/connection.go:26` - Database connection error
- `backend-go/workers/statistics_calculator.go:66,76,108,113,143,153,196,201` - Statistics calculation errors

## Key Conventions
- Always use `fmt.Errorf` with `%w` verb to wrap errors
- Provide descriptive context about what operation failed
- Include relevant identifiers (worker names, IDs, etc.) in error messages
- Preserve the original error chain for debugging
- Use consistent error message format: "failed to [action]: %w"
- Don't expose internal details in user-facing error messages
- Log wrapped errors with appropriate context for debugging
- Handle errors at appropriate levels - some should bubble up, others should be handled locally