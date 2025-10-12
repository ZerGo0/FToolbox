# Go Worker Pattern

## When to Use
Use this pattern for background tasks that need to run periodically, such as data synchronization, rank calculations, or statistics updates.

## Why It Exists
This pattern provides a robust framework for managing concurrent background workers with proper lifecycle management, error handling, and status tracking.

## Implementation
Workers implement the `Worker` interface and are managed by a `WorkerManager`. Key characteristics:

- Interface-based design with `Name()`, `Interval()`, and `Run()` methods
- BaseWorker struct for common functionality
- Manager handles registration, start/stop, and graceful shutdown
- Database-persisted worker status with failure counts and error tracking
- Panic recovery with stack trace logging
- Context-based cancellation for clean shutdown

## References
- `backend-go/workers/worker.go:8-13` - Worker interface definition
- `backend-go/workers/worker.go:15-34` - BaseWorker implementation
- `backend-go/workers/manager.go:15-33` - WorkerManager struct with concurrency safety
- `backend-go/workers/manager.go:35-67` - Worker registration with database sync
- `backend-go/workers/manager.go:152-195` - Worker execution loop with panic recovery
- `backend-go/workers/manager.go:197-290` - Individual worker execution with status updates
- `backend-go/main.go:96-141` - Worker registration and startup in main

## Key Conventions
- Always implement the Worker interface for consistency
- Use BaseWorker for common name and interval functionality
- Register workers before starting them
- Handle panics and update database status on failures
- Use context for cancellation and timeout handling
- Log both success and failure with duration metrics