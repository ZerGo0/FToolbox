# AGENTS.md

This file provides guidance to Codex CLI when working with code in this repository.

## Project Overview
- Name: FToolbox
- Description: Full-stack web application for Fansly analytics and tooling, with a SvelteKit frontend in `frontend` and a Go API in `backend-go`.
- Key notes:
  - The authoritative database schema lives in `backend-go/models`, and startup applies it through `backend-go/database/migrate.go`.
  - The frontend is deployed with Cloudflare Pages and depends on `PUBLIC_API_URL` at build time.
  - `Taskfile.yml` is mainly a human workflow entrypoint; its watch tasks start dev servers.

## Repository-Wide Rules
- Use `/tmp` for build artifacts or other generated output if a build is explicitly required.
- Do not run `task watch-frontend` or `task watch-backend` unless the user explicitly asks; both start dev servers.
- When you change a specific subproject, run only that subproject's documented static checks unless the task requires broader verification.

## High-Level Architecture
- `frontend`: SvelteKit 2 + Svelte 5 single-page app with Cloudflare adapter; client code calls the backend over `/api`.
- `backend-go`: Fiber API with Zap logging and GORM against MariaDB/MySQL.
- `backend-go/workers`: background jobs registered through `WorkerManager`; worker status is stored in the `workers` table and surfaced via `/api/workers/status`.
- `backend-go/fansly`: shared Fansly client that centralizes auth, retries, and global rate limiting.

## Repository Workflow
- Frontend uses `pnpm` from `frontend`.
- Backend uses Go modules from `backend-go`.
- `task install` installs frontend packages and runs `go mod download` for Go modules, but there is no single repo-wide lint or typecheck task in `Taskfile.yml`.
- `.github/workflows/deploy-backend-go.yml` is the current deployment workflow and builds the backend Docker image from `backend-go`.

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
  - Use `PUBLIC_API_URL` for backend requests instead of hardcoded backend hosts.
- Useful files:
  - `frontend/package.json`
  - `frontend/eslint.config.js`
  - `frontend/svelte.config.js`
  - `frontend/wrangler.toml`

### backend-go
- Language: Go
- Framework/runtime: Fiber v2, GORM, Zap
- Package manager: Go modules
- Important packages: `github.com/gofiber/fiber/v2`, `gorm.io/gorm`, `gorm.io/driver/mysql`, `go.uber.org/zap`
- Checks:
  - `go fmt ./...`
  - `go vet ./...`
- Rules / conventions:
  - Keep schema changes aligned with `backend-go/models`, which feed `database.AutoMigrate`.
- Useful files:
  - `backend-go/go.mod`
  - `backend-go/main.go`
  - `backend-go/config/config.go`
  - `backend-go/routes/routes.go`

## Key Architectural Patterns
- Backend routes are grouped under `/api`.
- Route-level failures should surface as JSON objects with an `error` string.
- Frontend API calls are expected to target absolute URLs built from `PUBLIC_API_URL`.
- Background jobs should use the existing worker manager flow instead of introducing separate recurring job mechanisms.

- **ALWAYS** at the end of your turn, ask a follow-up question for the next logical step (**DON'T** ask questions like "Should I run tests?" or "Should I lint?", only ask questions that are relevant to the task at hand)
