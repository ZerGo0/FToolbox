# Backend Logging Pattern

## When to Use
Use this pattern for all logging operations in the Go backend to maintain consistent structured logging across handlers, workers, and application lifecycle events.

## Why It Exists
Ensures consistent structured logging with proper error handling, context preservation, and performance monitoring throughout the application. Centralizes logging configuration and provides searchable, filterable log output.

## Implementation
Always use the global zap logger instance `zap.L()` for logging operations:

### Error Logging
```go
zap.L().Error("descriptive message", zap.Error(err))
```

### Info Logging with Context
```go
zap.L().Info("operation completed", zap.String("key", "value"), zap.Int("count", total))
```

### Debug Logging
```go
zap.L().Debug("detailed diagnostic info", zap.Any("data", complexObject))
```

## Source References
- `backend-go/workers/manager.go:152` - Worker lifecycle logging
- `backend-go/workers/creator_updater.go:45` - Error logging in worker operations
- `backend-go/handlers/tag_handler.go:188` - Request logging with context
- `backend-go/handlers/creator_handler.go:120` - Error logging in API handlers
- `backend-go/main.go:89` - Application startup logging
- `backend-go/fansly/client.go:347` - External API call logging

## Key Conventions
- Always use `zap.L()` for the global logger instance
- Include `zap.Error(err)` for error contexts
- Use structured fields (`zap.String`, `zap.Int`, `zap.Any`) instead of formatted strings
- Provide descriptive, action-oriented messages
- Log at appropriate levels (Error for failures, Info for significant events, Debug for diagnostics)