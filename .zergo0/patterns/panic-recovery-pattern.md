# Panic Recovery Pattern

## When to Use
Use this pattern for critical operations that could panic and need graceful error handling with proper cleanup and status tracking.

## Why It Exists
Provides robust error handling for operations that might panic due to unexpected conditions. Ensures proper cleanup, status updates, and logging when panics occur, preventing application crashes.

## Implementation

### Worker Panic Recovery
Recover from panics in worker operations with comprehensive error tracking:
```go
defer func() {
    if r := recover(); r != nil {
        zap.L().Error("Worker panicked",
            zap.String("worker", worker.Name()),
            zap.Any("panic", r),
            zap.String("stack", string(debug.Stack())))
        
        // Update database status
        updates := map[string]interface{}{
            "status":        "failed",
            "failure_count": gorm.Expr("failure_count + 1"),
            "last_error":    fmt.Sprintf("Worker panic: %v", r),
        }
    }
}()
```

### Transaction Panic Recovery
Recover from panics during database transactions with rollback:
```go
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        zap.L().Error("Transaction panicked",
            zap.Any("panic", r),
            zap.String("stack", string(debug.Stack())))
    }
}()
```

## Source References
- `backend-go/workers/manager.go:157-170` - Worker execution panic recovery with status updates
- `backend-go/workers/manager.go:240-253` - Secondary worker panic recovery
- `backend-go/workers/statistics_calculator.go:54-67` - Statistics calculation panic recovery with transaction rollback
- `backend-go/workers/statistics_calculator.go:131-144` - Creator statistics panic recovery with transaction rollback

## Key Conventions
- Always use `defer func()` with `recover()` for panic handling
- Log panics with worker context and stack traces using zap
- Update database status to "failed" when panics occur
- Increment failure counters atomically using `gorm.Expr`
- Rollback transactions immediately when panics occur
- Include panic details in error messages for debugging
- Use `debug.Stack()` to capture full stack traces
- Handle panics at appropriate levels - worker operations and critical transactions