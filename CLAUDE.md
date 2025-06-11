# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FToolbox is a full-stack web application that provides a collection of tools for Fansly creators and users. It features a SvelteKit frontend with real-time data visualization and a Bun backend with automated worker processes for data collection.

## Development Commands

### Frontend (SvelteKit)

```bash
cd frontend
pnpm check      # Run svelte-check for type errors
pnpm lint       # Run Prettier and ESLint
```

**ALWAYS** run `pnpm check && pnpm lint` after making changes

### Backend (Bun)

```bash
cd backend
bun check       # Run TypeScript type checking (tsc --noEmit)
bun lint        # Run Prettier and ESLint
```

**ALWAYS** run `bun check && bun lint` after making changes

### Task Runner

```bash
task watch-frontend  # Kill existing process and start frontend dev server
task watch-backend   # Kill existing process and start backend dev server
```

## Database Management

### Schema Location

- Schema: `/backend/src/db/schema.ts`
- Migrations: `/backend/drizzle/`

### Migration Commands

**IMPORTANT**: Always use Drizzle to generate migrations. NEVER write manual SQL migration files.

```bash
cd backend
bun drizzle-kit generate  # Generate migration after schema changes
```

Migrations run automatically on server startup via `runMigrations()` in index.ts.

### Database Tables

- `tags`: Fansly tags with view counts and metadata
- `tagHistory`: Historical view count tracking
- `tagRequests`: Queue for new tag addition requests
- `workers`: Worker status and run statistics

## Worker System Architecture

The application uses a sophisticated background worker system for data collection:

### Worker Manager

- Central `WorkerManager` class orchestrates all workers
- Prevents concurrent runs of the same worker
- Tracks worker status in database
- Supports graceful shutdown
- Environment control: `WORKER_ENABLED=false` disables all workers

### Available Workers

1. **tag-updater**: Updates view counts for all tracked tags (24-hour interval)
2. **tag-discovery**: Discovers new tags from Fansly
3. **rank-calculator**: Calculates global ranks based on view counts

### Worker Interface

```typescript
interface Worker {
  name: string;
  interval: number; // milliseconds
  run(): Promise<void>;
}
```

### Worker Status API

- Endpoint: `GET /api/workers/status`
- Returns: running/idle/failed status

## API Architecture

### Backend Stack

- **Framework**: Hono for HTTP server
- **Runtime**: Bun with TypeScript
- **Database**: SQLite with Drizzle ORM
- **CORS**: Enabled for frontend development

### API Patterns

- RESTful design with `/api/*` routes
- JSON request/response format
- Pagination, sorting, and filtering on list endpoints
- Proper HTTP status codes for errors
- Rate limiting considerations for external API calls

### Key Endpoints

- `GET /api/tags` - List tags with pagination/filtering
- `GET /api/tags/:name` - Get single tag details
- `POST /api/tags/request` - Request new tag tracking
- `GET /api/tags/:name/history` - Get tag history
- `GET /api/workers/status` - Worker system status

## Frontend Architecture

### Tech Stack

- **Framework**: SvelteKit with Svelte 5
- **Styling**: Tailwind CSS v4
- **Components**: shadcn-svelte (40+ UI components)
- **Deployment**: Cloudflare adapter

### Key Features

- File-based routing in `/src/routes/`
- Svelte 5 features including `.svelte.ts` files
- Pre-built UI components in `/src/lib/components/ui/`
- Responsive design with mobile considerations

## External API Integration

### Fansly API Client

- Location: `/backend/src/fansly/client.ts`
- Rate limiting implemented
- Error handling for API failures
- Tag discovery and view count fetching
