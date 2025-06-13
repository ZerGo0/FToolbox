# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FToolbox is a full-stack web application that provides a collection of tools for Fansly creators and users. The project is currently transitioning from a Bun/SQLite backend to a Go/MariaDB backend. Both backends coexist during the migration period.

## Current Architecture

- **Frontend**: SvelteKit with Svelte 5, Tailwind CSS v4, and shadcn-svelte components
- **Legacy Backend**: Bun runtime with Hono framework and SQLite database (port 3000)
- **New Backend**: Go with Fiber framework and MariaDB database (port 3000)

## Development Commands

### Frontend (SvelteKit)

```bash
cd frontend
pnpm check      # Run svelte-check for type errors
pnpm lint       # Run Prettier and ESLint
```

**ALWAYS** run `pnpm check && pnpm lint` after making changes

### Backend (Go)

**ALWAYS**: Run `go fmt ./...` and `go vet ./...` manually.

### Legacy Bun Backend

```bash
cd backend
bun check       # Run TypeScript type checking (tsc --noEmit)
bun lint        # Run Prettier and ESLint
```

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

### Legacy Bun Backend (SQLite)

- **Schema**: `/backend/src/db/schema.ts`
- **Migrations**: `/backend/drizzle/`
- **Migration command**: `cd backend && bun drizzle-kit generate`

### Database Tables (Both Backends)

- `tags`: Fansly tags with view counts and metadata
- `tag_histories`: Historical view count tracking (note: plural in Go backend)
- `tag_requests`: Queue for new tag addition requests
- `workers`: Worker status and run statistics

### Timestamp Handling

- **Bun Backend**: Unix timestamps (seconds since epoch) with automatic Date conversion
- **Go Backend**: Standard MySQL datetime fields, handled as `time.Time` in Go

## Worker System Architecture

Both backends implement the same worker system with different concurrency models:

### Workers

1. **tag-updater**: Updates view counts for all tracked tags
2. **tag-discovery**: Discovers new tags from Fansly
3. **rank-calculator**: Calculates global ranks (Go uses DB triggers)

### Implementation Differences

- **Bun**: Timer-based intervals with WorkerManager class
- **Go**: Goroutines with context-based cancellation and time.Ticker

### Configuration

- Environment variable: `WORKER_ENABLED=false` disables all workers
- Worker status endpoint: `GET /api/workers/status`

## API Architecture

### Common Endpoints (Both Backends)

- `GET /api/tags` - List tags with pagination/filtering
- `GET /api/tags/:name` - Get single tag details
- `POST /api/tags/request` - Request new tag tracking
- `GET /api/tags/:name/history` - Get tag history
- `GET /api/workers/status` - Worker system status
- `GET /` - Health check (Go backend)

### Backend Comparison

| Feature    | Bun Backend | Go Backend       |
| ---------- | ----------- | ---------------- |
| Port       | 3000        | 3000             |
| Framework  | Hono        | Fiber v2         |
| Database   | SQLite      | MariaDB          |
| ORM        | Drizzle     | GORM             |
| Logging    | Custom      | Zap (structured) |
| Hot Reload | bun --watch | Air              |

## Environment Configuration

### Go Backend (.env)

```env
# Database
DB_HOST=localhost
DB_PORT=50324
DB_USERNAME=root
DB_PASSWORD=devpassword
DB_DATABASE=ftoolbox

# Server
PORT=3000
LOG_LEVEL=debug

# Workers
WORKER_ENABLED=true

# External APIs
FANSLY_AUTH_TOKEN=
```

## External API Integration

### Fansly API Client

- **Bun**: `/backend/src/fansly/client.ts`
- **Go**: `/backend-go/fansly/client.go`
- Both implement rate limiting and error handling
- Tag discovery and view count fetching

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
- API calls to backend (configure port based on which backend is active)

## Important Migration Notes

1. The Go backend is the primary backend going forward
2. Frontend may need API endpoint updates if switching between backends
3. Database migration from SQLite to MariaDB requires data migration scripts
4. Worker intervals and configurations should match between backends
5. CORS is configured in both backends for frontend development
