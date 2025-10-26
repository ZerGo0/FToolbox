# AGENTS.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
- Name: FToolbox
- Description: Full-stack web application delivering analytics and tooling for Fansly creators and consumers with a SvelteKit frontend and Go backend.
- Key notes or warnings:
  - Database schema lives in `backend-go/models`, with AutoMigrate applying changes on startup; update models instead of issuing manual migrations.
  - Server listens on `PORT` (default `3000` from `backend-go/config/config.go`); the Taskfile only clears `3001` for the Air loop, so bind new services to `PORT`.
  - Tag heat is deprecated and all responses force `heat = 0`; never surface or depend on legacy heat values in new logic.
  - Fansly API calls must route through `backend-go/fansly/client.go`, which enforces global rate limiting and automatic 429 retries controlled by `FANSLY_GLOBAL_RATE_LIMIT` and `FANSLY_GLOBAL_RATE_LIMIT_WINDOW`.
  - Cloudflare Pages builds ignore the `wrangler.toml [vars]` block; configure `PUBLIC_API_URL` and other frontend env vars in the Pages dashboard before builds succeed.
  - Fansly auth: `FANSLY_AUTH_TOKEN` is read by the client and used when present.

## Global Rules
- **NEVER** use emojis!
- **NEVER** try to run the dev server!
- **NEVER** try to build in the project directory, always build in the `/tmp` directory!
- **NEVER** use comments in code - code should be self-explanatory
- **NEVER** cut corners, don't leave comments like `TODO: Implement X in the future here`! Always fully implement everything!
- **ALWAYS** when you're done, self-critique your work until you're sure it's correct
- **ALWAYS** at the end of your turn, ask a follow-up question for the next logical step (**DON'T** ask questions like "Should I run tests?" or "Should I lint?", only ask questions that are relevant to the task at hand)
 

## High-Level Architecture
- **Frontend:** SvelteKit 2 (Svelte 5) with Vite, Tailwind CSS v4, and shadcn-svelte; loaders build absolute API URLs using `PUBLIC_API_URL`.
- **Backend:** Go 1.25 Fiber API using Zap logging, GORM (MySQL driver) targeting MariaDB, shared middleware for compression, CORS, ETag, and request limiting, serving JSON under `/api/*`.
- **Workers:** Goroutine workers managed by `backend-go/workers/manager.go`; enablement and intervals come from env vars and statuses persist in the `workers` table.
- **Fansly Integration:** `backend-go/fansly/client.go` centralizes authentication, retries, and global rate limiting for all Fansly requests.
- **Schema Source of Truth:** `backend-go/models` combined with `backend-go/database/migrate.go` drives AutoMigrate; never mutate the database outside this layer.
- **Infrastructure Utilities:** `Taskfile.yml` coordinates human dev loops (e.g., port cleanup) but must not be invoked from the agent.

## Project Guidelines

### frontend
- Language: TypeScript
- Framework/runtime: SvelteKit 2 (Svelte 5) on Vite with Cloudflare adapter
- Package manager: pnpm
- Important Packages: `@sveltejs/kit`, `@sveltejs/adapter-cloudflare`, `tailwindcss@4`, `bits-ui`, `svelte-sonner`, `chart.js`, `chartjs-adapter-date-fns`
- Checks:
  - Syntax Check: `pnpm check`
  - Lint: `pnpm lint`
  - Format: `pnpm format`
  - **ALWAYS** run these after you are done making changes
- Rules / conventions:
  - **ALWAYS** build API requests with `$env/static/public` values (notably `PUBLIC_API_URL`) and emit absolute URLs from loaders or server-only modules.
  - **ALWAYS** compose UI from shadcn-svelte primitives in `frontend/src/lib/components/ui` before introducing new components.
  - **NEVER** nest interactive children inside trigger components; rely on the `local/no-nested-interactive` helpers.
  - **NEVER** start dev loops, previews, or automated tests; only run the listed checks when a human explicitly requests them.
  - **ALWAYS** keep Cloudflare Pages deployment metadata in sync between `frontend/wrangler.toml` and the Pages dashboard env configuration.
- Useful files:
  - `frontend/eslint.config.js`
  - `frontend/eslint-plugin-local/index.js`
  - `frontend/svelte.config.js`
  - `frontend/wrangler.toml`
  - `frontend/.env.example`

### backend-go
- Language: Go (1.25+)
- Framework/runtime: Fiber v2 with Zap logging
- Package manager: Go modules
- Important Packages: `github.com/gofiber/fiber/v2`, `gorm.io/gorm`, `gorm.io/driver/mysql`, `go.uber.org/zap`, `github.com/joho/godotenv`
- Checks:
  - Format: `go fmt ./...`
  - Vet: `go vet ./...`
  - **ALWAYS** run these after you are done making changes
- Rules / conventions:
  - **ALWAYS** call `config.Load()` and pass the resulting config explicitly so default intervals and rate limits apply.
  - **ALWAYS** persist schema changes in `backend-go/models` and let AutoMigrate handle DDL.
  - **ALWAYS** return `{ "error": string }` JSON responses and log diagnostics with `zap.L()` instead of exposing internals.
  - **ALWAYS** use the shared `fansly.Client` and global limiters for external API calls.
  - **NEVER** spin up worker loops, Air, or other dev tooling from this agent; restrict work to code edits.
  - **NEVER** run automated tests beyond `go fmt` and `go vet` unless a human explicitly asks.
- Useful files:
  - `backend-go/config/config.go`
  - `backend-go/database/connection.go`
  - `backend-go/database/migrate.go`
  - `backend-go/routes/routes.go`
  - `backend-go/fansly/client.go`
  - `backend-go/workers/manager.go`
  - `backend-go/utils/utils.go`

### root
- Language: Mixed (YAML, shell)
- Framework/runtime: Taskfile helpers and repo automation
- Package manager: pnpm (frontend), Go modules (backend)
- Important Packages: `task`, `pnpm`, `air`
- Checks:
  - None; root automation targets human workflows only.
  - **ALWAYS** run these after you are done making changes
- Rules / conventions:
  - **ALWAYS** use `/tmp` for any build artifacts generated by the agent.
  - **ALWAYS** treat `Taskfile.yml` and the `claude-code-github-bot-*.sh` scripts as references only.
  - **NEVER** invoke `task watch-frontend`, `task watch-backend`, or other commands that start dev servers.
  - **NEVER** alter automation scripts unless matching established patterns exactly.
  - **NEVER** hunt for or execute additional test suites; only run the documented checks when specifically instructed.
- Useful files:
  - `Taskfile.yml`
  - `claude-code-github-bot-setup.sh`
  - `claude-code-github-bot-cleanup.sh`

## Key Architectural Patterns
- **Error Handling:** Central Fiber error handler with Zap logging; route handlers return `{ "error": string }` payloads and avoid leaking internal details.
- **Request Shaping:** Pagination, sorting, and time-window helpers in `backend-go/handlers/tag_handler.go` define reusable query parsing patterns for list endpoints.
- **Rate Limiting:** Global middleware covers all requests, while sensitive POST routes apply stricter per-IP limiters—reuse the existing limiter helpers when adding routes.
- **Ranking & Statistics:** Ranking utilities reside in `backend-go/utils/utils.go` and run through workers scheduled by `WorkerManager`; never spawn ad-hoc goroutines.
- **Frontend Data Fetching:** Prefer SvelteKit loaders for initial data. Components may call the API using `PUBLIC_API_URL` from `$env/static/public`; never import non-public env values into client code.
- **UI Composition:** Tailwind utility patterns and shadcn-svelte primitives in `frontend/src/lib/components/ui` are the first choice for new UI; do not duplicate styling.
 
