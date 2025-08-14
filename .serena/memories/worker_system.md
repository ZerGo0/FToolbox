# Worker System Architecture

## Overview
Background job system implemented in Go using goroutines with context-based cancellation and time.Ticker for scheduling.

## Core Workers

### 1. Tag Updater (`tag-updater`)
- **Purpose**: Updates view counts for all tracked tags
- **Interval**: Configurable (default: hourly)
- **Process**: Fetches latest view counts from Fansly API and updates database

### 2. Creator Updater (`creator-updater`)
- **Purpose**: Updates subscriber counts for all tracked creators
- **Interval**: Configurable (default: hourly)
- **Process**: Fetches latest subscriber data from Fansly API

### 3. Tag Discovery (`tag-discovery`)
- **Purpose**: Discovers new tags from Fansly
- **Interval**: Configurable (default: daily)
- **Process**: Searches for new tags and adds them to tracking

### 4. Rank Calculator (`rank-calculator`)
- **Purpose**: Calculates tag and creator rankings
- **Interval**: After each update cycle
- **Process**: Sorts by view/subscriber count and assigns ranks

### 5. Statistics Calculator (`statistics-calculator`)
- **Purpose**: Calculates heat scores and trends
- **Interval**: After rank calculation
- **Process**: Analyzes historical data for trends

## Implementation Details

### Worker Manager
- Location: `/backend-go/workers/manager.go`
- Manages worker lifecycle (start, stop, status)
- Coordinates worker execution
- Handles graceful shutdown

### Worker Features
- **Concurrency**: Each worker runs in its own goroutine
- **Cancellation**: Context-based cancellation for graceful shutdown
- **Scheduling**: time.Ticker for regular intervals
- **Status Tracking**: Database persistence of worker status
- **Error Handling**: Retry logic with exponential backoff
- **Rate Limiting**: Respects Fansly API rate limits

### Configuration
Workers can be enabled/disabled and intervals configured via environment variables:
- `WORKER_TAG_UPDATER_ENABLED`
- `WORKER_TAG_UPDATER_INTERVAL`
- Similar patterns for other workers

### Monitoring
- Worker status available via `/api/workers/status` endpoint
- Tracks: last_run, next_run, run_count, error_count
- Logs detailed execution information via Zap logger