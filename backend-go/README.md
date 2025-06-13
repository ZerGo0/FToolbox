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
- `GET /api/workers/status` - Worker system status
- `GET /api/health` - Health check

## Technologies

- **Fiber** - Web framework
- **GORM** - ORM with auto-migration
- **Zap** - Structured logging
- **Air** - Hot reloading
- **MariaDB** - Database