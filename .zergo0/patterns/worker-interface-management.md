# Worker Interface and Management Pattern

## When to Use
Use this pattern for any background worker that needs to run periodically with proper lifecycle management and error handling.

## Why It Exists
This pattern provides consistent worker behavior, panic recovery, status tracking, and graceful shutdown capabilities across all background tasks.

## Implementation Details

### Worker Interface
All workers must implement the Worker interface:
```go
type Worker interface {
    Name() string
    Interval() time.Duration
    Run(ctx context.Context) error
}
```

### BaseWorker Pattern
Use BaseWorker for common functionality:
```go
type BaseWorker struct {
    name     string
    interval time.Duration
}

func NewBaseWorker(name string, interval time.Duration) BaseWorker {
    return BaseWorker{name: name, interval: interval}
}
```

### Worker Registration
Register workers with the manager:
```go
workerManager.Register(NewSpecificWorker(config))
```

### Panic Recovery
Workers include comprehensive panic recovery:
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

### Status Tracking
Update worker status in database:
```go
updates := map[string]interface{}{
    "status":      "running",
    "last_run_at": time.Now(),
    "next_run_at": time.Now().Add(worker.Interval()),
}

if err != nil {
    updates["status"] = "failed"
    updates["failure_count"] = gorm.Expr("failure_count + 1")
    updates["last_error"] = err.Error()
} else {
    updates["success_count"] = gorm.Expr("success_count + 1")
}
```

## References
- `backend-go/workers/worker.go:8-35` - Worker interface definition
- `backend-go/workers/manager.go:152-290` - Worker execution with panic recovery
- `backend-go/workers/manager.go:35-67` - Worker registration pattern