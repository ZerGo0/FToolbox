# FToolbox Backend (Go)

Go implementation of the FToolbox backend API.

## Prerequisites

- Go 1.21+
- MariaDB
- Air (for hot reloading): `go install github.com/cosmtrek/air@latest`

## Setup

1. Copy `.env.example` to `.env` and configure your database settings
2. Install dependencies: `go mod download`
3. Run with hot reloading: `air`
4. Or run directly: `go run main.go`

## API Endpoints

- `GET /api/tags` - List tags with pagination/filtering
- `GET /api/tags/:name` - Get single tag details
- `POST /api/tags/request` - Request new tag tracking
- `GET /api/tags/:name/history` - Get tag history
- `GET /api/tags/related` - Get related tags
- `GET /api/workers/status` - Worker system status
- `GET /api/health` - Health check

## Technologies

- **Fiber** - Web framework
- **GORM** - ORM with auto-migration
- **Zap** - Structured logging
- **Air** - Hot reloading
- **MariaDB** - Database

## Related Tags Endpoint

`GET /api/tags/related`

Query params:

- `tags` (required): comma-separated tag names to base the recommendations on.
- `limit` (optional): number of results to return (default 10, max 20).
- `mode` (optional): scoring mode. Supported: `smart` (default), `popular`.
- `windowDays` (optional): lookback window in days. Default 14, clamped to [7, 30].
- `minViewCount` (optional): minimum `view_count` for candidate tags. Default 5000.
- `minCoverage` (optional, smart mode): minimum number of input tags a candidate must co-occur with. Default is ceil(40% of inputs).

Responses include per-tag fields:

- `id`, `tag`
- `score`: numeric score (equals `finalScore` in smart mode)
- `normScore`, `coverage`, `finalScore` (smart mode only)

Top-level metadata:

- `source`: `computed` for smart mode, `precomputed` for popular
- `mode`, `windowDays`, `minViewCount`, `minCoverage`, `usedTagIds`
