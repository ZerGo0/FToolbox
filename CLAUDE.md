# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
- Name: FToolbox
- Description: Full‑stack web application providing tools and insights for Fansly creators and users. Frontend is SvelteKit; backend is a Go API with background workers and a MariaDB database.
- Key notes/warnings:
  - Database schema is defined by GORM models in `backend-go/models` and applied via AutoMigrate on startup. Changing models migrates the DB automatically.
  - Backend listen port is `PORT` (defaults to `3000` per `backend-go/config/config.go`). The root Taskfile stops `3001` for the Air dev loop; rely on `PORT` for the actual server port.
  - Tag “heat” is deprecated and forced to `0` in responses; do not build UI logic that depends on heat.
  - Fansly API access uses a global rate limiter with simple retries and special handling for HTTP 429. Configure with `FANSLY_GLOBAL_RATE_LIMIT` and `FANSLY_GLOBAL_RATE_LIMIT_WINDOW`; optional `FANSLY_AUTH_TOKEN` adds Authorization.
  - Frontend environment variables (e.g., `PUBLIC_API_URL`) are needed at build time on Cloudflare Pages; set them in the Pages dashboard. Do not assume `wrangler.toml [vars]` will be picked up by SvelteKit during build.
  - Worker intervals are configured via env: `WORKER_UPDATE_INTERVAL`, `WORKER_DISCOVERY_INTERVAL`, `RANK_CALCULATION_INTERVAL`, `WORKER_STATISTICS_INTERVAL`.

## Global Rules

- **NEVER** use emojis!
- **NEVER** try to run the dev server!
- **NEVER** try to build in the project directory, always build in the `/tmp` directory!
- **ALWAYS** search for existing code patterns in the codebase and follow them consistently
- **NEVER** use comments in code — code should be self-explanatory

## High-Level Architecture
- Frontend: SvelteKit 2 (Svelte 5), Tailwind CSS v4, shadcn‑svelte components checked into `frontend/src/lib/components/ui`, deployed on Cloudflare Pages. Uses `PUBLIC_API_URL` to call the API.
- Backend: Go 1.25+, Fiber v2, GORM (MySQL driver), Zap logging, MariaDB. API routes under `/api/*` with JSON responses and request limiting.
- Workers: Goroutine-based workers managed by a `WorkerManager` with DB‑persisted status in `workers` table; intervals and enablement via env (`WORKER_ENABLED`, `*_INTERVAL`).
- Schema Source of Truth: `backend-go/models` (GORM models) with AutoMigrate in `backend-go/database/migrate.go`.

## Project Guidelines

### frontend
- Language: TypeScript
- Framework/runtime: SvelteKit 2 (Svelte 5), Vite
- Package manager: pnpm
- Important Packages: `@sveltejs/kit`, `@sveltejs/adapter-cloudflare`, `tailwindcss@4`, `bits-ui`, `vaul-svelte`, `svelte-sonner`
- Checks:
  - Syntax Check: `pnpm check`
  - Lint: `pnpm lint`
  - Format: `pnpm format`
  - **ALWAYS** run these after you are done making changes
- Rules/conventions:
  - **ALWAYS** use shadcn‑svelte components from `frontend/src/lib/components/ui` where possible.
  - **NEVER** nest interactive components inside Trigger components (custom ESLint rule `local/no-nested-interactive`). Use the child‑snippet pattern or apply `class={buttonVariants(...)}` on the Trigger instead.
  - **ALWAYS** read the API base from `$env/static/public` (`PUBLIC_API_URL`) and build absolute fetch URLs.
- Useful files:
  - `frontend/eslint.config.js` — enables the custom rule.
  - `frontend/eslint-plugin-local/index.js` — `no-nested-interactive` rule definition.
  - `frontend/wrangler.toml` — Pages build command and routes (env must still be set in the Pages dashboard for SvelteKit).
  - `frontend/.env.example` — `PUBLIC_API_URL` sample.

### backend-go
- Language: Go (1.25+)
- Framework/runtime: Fiber v2, Zap, GORM (MySQL), MariaDB
- Package manager: Go modules
- Important Packages: `github.com/gofiber/fiber/v2`, `gorm.io/gorm`, `gorm.io/driver/mysql`, `go.uber.org/zap`, `github.com/joho/godotenv`
- Checks:
  - Format: `go fmt ./...`
  - Vet: `go vet ./...`
  - **ALWAYS** run these after you are done making changes
- Rules/conventions:
  - **ALWAYS** use the global Zap logger (`zap.L()`) and return JSON errors as `{ "error": string }` without exposing internal errors.
  - **ALWAYS** let AutoMigrate manage schema changes; edit GORM models in `backend-go/models` instead of hand‑editing tables.
  - **NEVER** depend on “heat” — it is set to `0` in responses by design.
  - **ALWAYS** honor Fansly global rate limiting when adding client calls.
- Useful files:
  - `backend-go/config/config.go` — environment config and defaults (e.g., `PORT`, rate limits, worker intervals).
  - `backend-go/database/connection.go` and `backend-go/database/migrate.go` — DB connection and AutoMigrate.
  - `backend-go/fansly/client.go` — HTTP client with global rate limiter and retry/backoff.
  - `backend-go/routes/routes.go` and `backend-go/handlers/*` — API surface and behavior.
  - `backend-go/workers/*` and `backend-go/models/worker.go` — worker manager, workers, and persisted status.
  - `backend-go/.env.example` — complete backend env reference.
  - `backend-go/.air.toml` — Air dev loop configuration.

### root
- Task Runner: Taskfile with helpers for local loops
  - `task watch-frontend`
  - `task watch-backend`

## Key Architectural Patterns
- Error handling: Central Fiber error handler returns `{ error }` JSON; handlers log via Zap and avoid leaking internals.
- Database access: GORM models + AutoMigrate; ranking updates done via SQL helpers in `backend-go/utils/utils.go`.
- Request shaping: Query param mapping in handlers; pagination (`page`, `limit`) with sane clamps; explicit sorting maps to DB columns; history includes optional windows.
- Rate limiting: Global Fansly limiter in client; per‑route request limiting middleware for sensitive endpoints.
