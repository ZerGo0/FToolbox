# AGENTS.md

This file provides guidance to Codex CLI when working with code in this repository.

## Project Overview
- Name: FToolbox
- Description: Full-stack web application for Fansly analytics and tooling, with a SvelteKit frontend in `frontend` and a Go API in `backend-go`.
- Key notes:
  - The authoritative database schema lives in `backend-go/models`, and startup applies it through `backend-go/database/migrate.go`.
  - The checked-in frontend deploy path is `frontend/wrangler.toml` plus `pnpm deploy:production`; `frontend/README.md` still mentions Cloudflare Pages, so prefer the config and package scripts when they conflict.
  - `Taskfile.yml` is a convenience entrypoint for humans; the watch tasks start dev servers and are not general-purpose verification commands.

## Repository-Wide Rules
- Use `/tmp` for build artifacts or other generated output if a build is explicitly required.
- Do not run `task watch-frontend` or `task watch-backend` unless the user explicitly asks; both start dev servers.
- When you change a specific subproject, run only that subproject's documented static checks unless the task requires broader verification.
- End each turn with a follow-up question about the next logical repository task.

## High-Level Architecture
- `frontend`: SvelteKit 2 + Svelte 5 client application. `frontend/src/routes/+layout.ts` disables SSR, so route loads and components call the backend with absolute URLs built from `PUBLIC_API_URL`.
- `backend-go`: Fiber API with Zap logging, GORM models, and a MariaDB/MySQL connection layer.
- `backend-go/workers`: recurring jobs registered through `workers.WorkerManager`; worker state is persisted in the `workers` table and exposed through `/api/workers/status`.
- `backend-go/fansly`: shared Fansly client that centralizes auth, retries, and global rate limiting.

## Repository Workflow
- Frontend uses `pnpm` from `frontend`.
- Backend uses Go modules from `backend-go`.
- `task install` runs `pnpm install -r` and `go mod download` for each Go module found in the repo; `task update` runs `pnpm update -r` plus `go get -u ./... && go mod tidy` in each Go module.
- The checked-in frontend deployment command is `pnpm deploy:production` from `frontend`.
- `.github/workflows/deploy-backend-go.yml` is the only checked-in CI/deploy workflow; it builds the backend Docker image from `backend-go` and rsyncs deploy artifacts to the server.

## Project Guidelines

### frontend
- Language: TypeScript
- Framework/runtime: SvelteKit 2, Svelte 5, Vite, Cloudflare adapter
- Package manager: `pnpm`
- Important packages: `@sveltejs/kit`, `@sveltejs/adapter-cloudflare`, `bits-ui`, `tailwindcss`, `sveltekit-superforms`, `chart.js`
- Checks:
  - `pnpm check`
  - `pnpm lint`
  - `pnpm format`
- Rules / conventions:
  - Route and component fetches should go straight to `${PUBLIC_API_URL}/api/...`; there is no checked-in SvelteKit server proxy layer.
  - Prefer the existing UI/component seams in `src/lib/components` and `src/lib/components/ui` instead of introducing parallel patterns.
- Useful files:
  - `frontend/package.json`
  - `frontend/eslint.config.js`
  - `frontend/svelte.config.js`
  - `frontend/wrangler.toml`
  - `frontend/src/routes/+layout.ts`

### backend-go
- Language: Go
- Framework/runtime: Fiber v2, GORM, Zap
- Package manager: Go modules
- Important packages: `github.com/gofiber/fiber/v2`, `gorm.io/gorm`, `gorm.io/driver/mysql`, `go.uber.org/zap`, `github.com/joho/godotenv`
- Checks:
  - `go fmt ./...`
  - `go vet ./...`
- Rules / conventions:
  - Keep schema changes aligned with `backend-go/models` and `backend-go/database/migrate.go`; this repo does not maintain a separate handwritten migration layer.
  - Route external Fansly calls through `backend-go/fansly` and recurring background work through `backend-go/workers`.
- Useful files:
  - `backend-go/go.mod`
  - `backend-go/main.go`
  - `backend-go/config/config.go`
  - `backend-go/routes/routes.go`
  - `backend-go/database/migrate.go`

## Key Architectural Patterns
- Backend routes live under `/api`, and handler failures are expected to surface as JSON objects with an `error` string.
- Frontend code fetches backend data from absolute URLs built with `PUBLIC_API_URL` rather than relative `/api` calls.
- GORM models plus `database.AutoMigrate` are the schema contract; add new tables there instead of introducing a separate migration system.
- Background jobs should use the existing worker manager flow instead of introducing separate scheduling/orchestration paths.

- **ALWAYS** at the end of your turn, ask a follow-up question for the next logical step (**DON'T** ask questions like "Should I run tests?" or "Should I lint?", only ask questions that are relevant to the task at hand)
