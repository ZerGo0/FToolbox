# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FToolbox is a full-stack web application that provides a collection of tools for Fansly creators and users.

## Current Architecture

- **Frontend**: SvelteKit with Svelte 5, Tailwind CSS v4, and shadcn-svelte components (deployed on Cloudflare Pages)
- **Backend**: Go 1.24+ with Fiber framework, GORM ORM, Zap logging, and MariaDB database (port 3000)
- **Task Runner**: Taskfile for development automation

## Development Commands

### Frontend (SvelteKit)

```bash
cd frontend
pnpm install    # Install dependencies
pnpm check      # Run svelte-check for type errors
pnpm lint       # Run Prettier and ESLint
```

**ALWAYS** run `pnpm check && pnpm lint` after making changes
**ALWAYS** use shadcn svelte components for UI if possible @frontend/src/lib/components/ui/

### Backend (Go)

```bash
cd backend-go
go mod download # Install dependencies
go fmt ./...    # Format code
go vet ./...    # Vet code for issues
```

**ALWAYS** run `go fmt ./...` and `go vet ./...` manually after changes

### Task Runner

```bash
task watch-frontend  # Kill port 5173 and start frontend dev server
task watch-backend   # Kill port 3000 and start Go backend with Air
```

## Database Management

### Go Backend (MariaDB)

- **Schema**: GORM models in `/backend-go/models/`
- **Connection**: Configured in `/backend-go/database/connection.go`
- **Migrations**: Auto-migration via GORM on startup
- **Manual migrations**: `/backend-go/database/sql_migrations.go` for complex operations

**Important**: The Go backend uses GORM's AutoMigrate feature. Database schema changes are applied automatically when models are modified.

### Database Tables

- `tags`: Fansly tags with view counts and metadata
- `tag_histories`: Historical view count tracking
- `tag_requests`: Queue for new tag addition requests
- `workers`: Worker status and run statistics

## Worker System Architecture

### Workers

1. **tag-updater**: Updates view counts for all tracked tags
2. **tag-discovery**: Discovers new tags from Fansly
3. **rank-calculator**: Calculates tag rankings

### Implementation

- Go implementation uses goroutines with context-based cancellation and time.Ticker
- Workers run on configurable intervals and can be enabled/disabled
- Worker status persisted in database

## API Architecture

### Main Endpoints

- `GET /api/tags` - List tags with pagination/filtering
- `GET /api/tags/:name` - Get single tag details
- `POST /api/tags/request` - Request new tag tracking
- `GET /api/tags/:name/history` - Get tag history
- `GET /api/workers/status` - Worker system status
- `GET /api/health` - Health check

## External API Integration

### Fansly API Client

- Location: `/backend-go/fansly/client.go`
- Implements global rate limiting across all endpoints
- Automatic retry with exponential backoff
- Tag discovery and view count fetching

### Rate Limiting Configuration

Environment variables:

- `FANSLY_GLOBAL_RATE_LIMIT=50` - Global rate limit (requests per window)
- `FANSLY_GLOBAL_RATE_LIMIT_WINDOW=10` - Global rate limit window (seconds)
- `FANSLY_AUTH_TOKEN` - Optional authentication token for Fansly API

Features:

- Global rate limiting to spread requests evenly across all endpoints
- Simple retry logic with exponential backoff for failed requests
- Special handling for rate limit errors (429 status code)

## Code Architecture Patterns

### Go Backend Structure

```
backend-go/
├── config/         # Environment configuration
├── database/       # DB connection and migrations
├── models/         # GORM models
├── handlers/       # HTTP request handlers
├── routes/         # Route definitions
├── workers/        # Background job implementations
└── fansly/         # External API client
```

### Frontend Patterns

- File-based routing in `/src/routes/`
- Svelte 5 runes and `.svelte.ts` files
- Pre-built UI components from shadcn-svelte
- API calls to backend at port 3000
- Custom ESLint rule to prevent nested interactive elements
