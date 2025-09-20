# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
- Name
  - FToolbox
- Description
  - Full‑stack web application offering tools for Fansly creators and users. Monorepo with a SvelteKit frontend and a Go API backend.
- Key notes or warnings (e.g. API quirks, data formats, environment caveats)
  - Frontend builds for Cloudflare Pages; `PUBLIC_API_URL` must be set at build time (Cloudflare Pages env). Do not rely on `[vars]` alone in `wrangler.toml` for SvelteKit envs.
  - Backend uses Go 1.25+ (see `backend-go/go.mod`), Fiber, GORM (AutoMigrate), Zap, and MariaDB/MySQL. Default port is `3000` (`PORT` env).
  - External Fansly API access is wrapped by `backend-go/fansly/client.go` with a global rate limiter; configure via `FANSLY_GLOBAL_RATE_LIMIT` and `FANSLY_GLOBAL_RATE_LIMIT_WINDOW`.
  - Workers are controlled by `WORKER_ENABLED` and persist status in DB. Do not assume ad‑hoc endpoints exist—check `backend-go/routes/routes.go` as the source of truth.

## Global Rules
The following rules MUST always be included verbatim in every CLAUDE.md:
- **NEVER** use emojis!
- **NEVER** try to run the dev server!
- **NEVER** try to build in the project directory, always build in the `/tmp` directory!
- **ALWAYS** search for existing code patterns in the codebase and follow them consistently
- **NEVER** use comments in code — code should be self-explanatory

Additional critical rules for this repo:
- **NEVER** try to run `git commit` unless explicitly told to!
- **ALWAYS** do what the rules say in AGENTS.md, **DON'T** ask for permissions just do what the rules say
- **NEVER** leave bread crumbs when you delete old code!
- **NEVER** inline imports

## High-Level Architecture
- Databases, services, frameworks, or core technologies
  - Backend: Go 1.25+, Fiber, GORM (MySQL driver), Zap, MariaDB/MySQL.
  - Frontend: SvelteKit 2, Svelte 5, Tailwind CSS v4, shadcn‑svelte components, Vite, Cloudflare adapter + Wrangler.
  - Rate limiting: app-level request limiter on selected routes; global Fansly API limiter in `backend-go/ratelimit/`.
- How different systems interact (backend, frontend, workers, etc.)
  - Frontend reads `PUBLIC_API_URL` and calls the backend under `/api/*`.
  - Backend exposes REST endpoints; background workers update DB on intervals and record status in `workers`.
  - Backend’s Fansly client performs rate‑limited HTTP calls with retry/backoff, then handlers shape JSON responses.
- Where the source of truth lives for schemas or shared types
  - Database schema: `backend-go/models` (GORM models) + `backend-go/database/migrate.go` (AutoMigrate set). This is authoritative.
  - API routes: `backend-go/routes/routes.go`.
  - API payload shapes: defined in handlers (`backend-go/handlers/*`). No generated shared TS types.

## Project Guidelines
For each major project in the monorepo:

- backend-go
  - Language(s) used
    - Go
  - Framework(s) / runtime
    - Fiber; Go 1.25+; Zap logging
  - Package manager
    - Go modules (`go.mod`)
  - Important Packages (list key dependencies, not exhaustive)
    - `github.com/gofiber/fiber/v2`, `gorm.io/gorm`, `gorm.io/driver/mysql`, `go.uber.org/zap`, `github.com/joho/godotenv`
  - Checks (list the syntax check and lint commands if available; Claude must run these every time after making changes)
    - `go fmt ./...` and `go vet ./...`
    - Install deps when needed: `go mod download`
    - Build only in `/tmp`: e.g., `GOOS=linux GOARCH=amd64 go build -o /tmp/ftoolbox-api` (do not build inside the repo)
  - Rules / conventions (prefer ALWAYS/NEVER phrasing for clarity)
    - **ALWAYS** add/modify tables via GORM models in `backend-go/models` (AutoMigrate applies changes).
    - **ALWAYS** wire new endpoints through `routes/routes.go` and a handler in `handlers/`—keep JSON error shape `{ "error": string }` consistent with `main.go` error handler.
    - **ALWAYS** use Zap for logs; prefer structured fields.
    - **NEVER** invent endpoints—verify in `routes/routes.go`.
    - **NEVER** bypass the global Fansly client; use `fansly.Client` with its limiter.
  - Useful files (only those that need calling out)
    - `backend-go/main.go` — app wiring, middleware, app limiter, worker startup
    - `backend-go/routes/routes.go` — canonical list of REST routes
    - `backend-go/handlers/` — `Tag`, `Creator`, `Worker` handlers and response shaping
    - `backend-go/models/` — DB schema (tags, tag_history, creators, creator_history, tag_statistics, creator_statistics, tag_relations_daily, workers)
    - `backend-go/database/connection.go`, `backend-go/database/migrate.go` — DB connect + AutoMigrate
    - `backend-go/fansly/client.go` — external API client with retries and rate limiting
    - `backend-go/ratelimit/global.go` — global limiter implementation
    - `backend-go/workers/` — `tag-updater`, `tag-discovery`, `rank-calculator`, `creator-updater`, `statistics-calculator`

- frontend
  - Language(s) used
    - TypeScript, Svelte 5
  - Framework(s) / runtime
    - SvelteKit 2 + Vite 6; Cloudflare adapter
  - Package manager
    - pnpm
  - Important Packages (list key dependencies, not exhaustive)
    - `@sveltejs/kit`, `svelte`, `@tailwindcss/vite`, `tailwindcss@^4`, `bits-ui`, `lucide-svelte`, `chart.js`, `wrangler`
  - Checks (list the syntax check and lint commands if available; Claude must run these every time after making changes)
    - `pnpm check`
    - `pnpm lint`
  - Rules / conventions (prefer ALWAYS/NEVER phrasing for clarity)
    - **ALWAYS** use existing shadcn‑svelte components from `frontend/src/lib/components/ui/` when possible.
    - **ALWAYS** read the API base from `PUBLIC_API_URL` (`$env/static/public`).
    - **ALWAYS** satisfy ESLint’s `local/no-nested-interactive` rule (see `frontend/eslint.config.js`).
    - **NEVER** hardcode backend URLs—compose `${PUBLIC_API_URL}/api/...`.
    - **NEVER** inline imports; keep module imports at top scope.
  - Useful files (only those that need calling out)
    - `frontend/package.json` — scripts (`check`, `lint`, `build`)
    - `frontend/svelte.config.js`, `frontend/svelte.check.config.js`
    - `frontend/eslint.config.js` — includes custom local plugin and rule
    - `frontend/wrangler.toml` — Cloudflare Pages config (build command sets `PUBLIC_API_URL`)
    - `frontend/.env.example` — shows `PUBLIC_API_URL`

- repo root
  - Task runner
    - `Taskfile.yml` — `watch-frontend`, `watch-backend` exist for humans; Claudia/agents must not run dev servers.
  - Formatting
    - `.prettierrc` (root) and front‑end Prettier config

## Key Architectural Patterns
- Shared conventions across components/modules (e.g. React/Svelte patterns, DB usage, error handling)
  - Error handling: Fiber global error handler returns `{ "error": string }`; route‑level limiters protect `/api/tags/request` and `/api/creators/request`.
  - Database: GORM models define columns and indexes; `AutoMigrate` runs at startup. History tables store time series; statistics tables aggregate 24h changes; `tag_relations_daily` stores co‑occurrence counts for recommendations.
  - Workers: context‑driven loops using `time.Ticker`; status persisted in `workers` with success/failure counters via `WorkerManager`.
  - Rate limiting: global client limiter spaces outbound Fansly requests; Fiber `limiter` middleware caps per‑IP requests on selected endpoints; app‑wide limiter in `main.go` (60/min) using `X-Forwarded-For`.
  - Frontend: API calls centralized on `PUBLIC_API_URL`; UI built from reusable shadcn‑svelte components; Tailwind v4 utility classes throughout.
