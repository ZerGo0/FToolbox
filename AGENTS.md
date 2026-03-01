# AGENTS.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
- Name: FToolbox
- Description: Full-stack web application delivering analytics and tooling for Fansly creators and consumers with a SvelteKit frontend and Go backend.
- Key notes or warnings:
  - Database schema source of truth is `backend-go/models`; startup runs AutoMigrate via `backend-go/database/migrate.go`.
  - Backend listens on `PORT` (default `3000` from `backend-go/config/config.go`); `Taskfile.yml` only clears `3001` for the Air loop.
  - Tag heat is legacy-only in this codebase; API responses currently force `heat = 0`.
  - Fansly API access must go through `backend-go/fansly/client.go`, which applies global throttling and retries.
  - Cloudflare Pages needs `PUBLIC_API_URL` configured in the Pages dashboard at build time; the `wrangler.toml [vars]` block is not sufficient for SvelteKit builds.

## Global Rules
- **NEVER** use emojis!
- **NEVER** try to run the dev server unless explicitly asked.
- **NEVER** try to build in the project directory; always build in the `/tmp` directory unless explicitly asked to build in the project directory.
- **NEVER** use comments in code - code should be self-explanatory
- **NEVER** cut corners, don't leave comments like `TODO: Implement X in the future here`! Always fully implement everything!
- **NEVER** revert/delete any changes that you don't know about! Always assume that we are in the middle of a task and that the changes are intentional!
- **ALWAYS** at the end of your turn, ask a follow-up question for the next logical step (**DON'T** ask questions like "Should I run tests?" or "Should I lint?", only ask questions that are relevant to the task at hand)
## Refactor Using Established Engineering Principles
After generating or editing code, you must always refactor your changes using well-established software engineering principles. These apply every time, without relying on diff inspection.

### Core Principles
- **DRY (Don’t Repeat Yourself)**: Eliminate duplicate or repetitive logic by consolidating shared behavior into common functions or helpers.  
- **KISS (Keep It Simple, Stupid)**: Prefer simple, straightforward solutions over unnecessarily complex or abstract designs.  
- **YAGNI (You Aren’t Gonna Need It)**: Only implement what is required for the current task; avoid speculative features or abstractions.

### Refactoring Requirements
1. Ensure the intent of your change is clear, explicit, and easy to understand.  
2. Maintain consistency with existing patterns, naming, and structure in the codebase.  
3. Remove duplication and merge similar logic following DRY.  
4. Simplify complex code paths or structures following KISS.  
5. Avoid adding features, hooks, or abstractions that the current task does not need, following YAGNI.  

**Principle:**  
> Every change must simplify the codebase, reduce duplication, clarify intent, and make the system easier to maintain.

## High-Level Architecture
- **Frontend:** SvelteKit 2 + Svelte 5 + Vite + Tailwind v4 + shadcn-svelte primitives (`frontend/src/lib/components/ui`). App runs with `ssr = false` in `frontend/src/routes/+layout.ts`.
- **Backend:** Go 1.26 Fiber API with Zap logging and GORM (MariaDB/MySQL driver), exposing JSON routes under `/api/*`.
- **Workers:** Background jobs are managed through `backend-go/workers/manager.go`; worker state is persisted in the `workers` table.
- **Fansly Integration:** `backend-go/fansly/client.go` owns auth header usage (`FANSLY_AUTH_TOKEN`), retries, and global throttling (`FANSLY_GLOBAL_RATE_LIMIT`, `FANSLY_GLOBAL_RATE_LIMIT_WINDOW`).
- **Schema Source of Truth:** GORM models in `backend-go/models` with startup AutoMigrate in `backend-go/database/migrate.go`.
- **Infrastructure:** `Taskfile.yml` exists for human workflows and starts dev loops; treat it as reference only in agent work.

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
  - **ALWAYS** build API requests with `$env/static/public` (`PUBLIC_API_URL`) and use absolute API URLs.
  - **NEVER** hardcode backend hosts like `http://localhost:3000` in frontend code.
  - **ALWAYS** compose UI from `frontend/src/lib/components/ui` primitives before adding new component patterns.
  - **NEVER** nest interactive components inside trigger components; the local ESLint rule `local/no-nested-interactive` enforces this.
  - **NEVER** run dev loops, previews, builds, or automated tests unless explicitly asked; run only the checks listed above.
  - **ALWAYS** run the listed checks after frontend changes.
  - **ALWAYS** keep Cloudflare Pages env settings aligned with frontend config for deployability.
- Useful files:
  - `frontend/eslint.config.js`
  - `frontend/eslint-plugin-local/index.js`
  - `frontend/svelte.config.js`
  - `frontend/wrangler.toml`
  - `frontend/.env.example`

### backend-go
- Language: Go (1.26+)
- Framework/runtime: Fiber v2 with Zap logging
- Package manager: Go modules
- Important Packages: `github.com/gofiber/fiber/v2`, `gorm.io/gorm`, `gorm.io/driver/mysql`, `go.uber.org/zap`, `github.com/joho/godotenv`
- Checks:
  - Format: `go fmt ./...`
  - Vet: `go vet ./...`
  - **ALWAYS** run these after you are done making changes
- Rules / conventions:
  - **ALWAYS** load configuration via `config.Load()` and pass config into worker/client setup.
  - **ALWAYS** implement schema changes in `backend-go/models` and rely on AutoMigrate.
  - **ALWAYS** return JSON errors in `{ "error": string }` shape for route-level failures.
  - **ALWAYS** use the shared `fansly.Client` for Fansly API calls so retry/rate-limit behavior stays centralized.
  - **NEVER** start Air/dev server/worker loops from agent tasks unless explicitly asked.
  - **NEVER** run automated tests unless explicitly asked; run only `go fmt ./...` and `go vet ./...`.
  - **ALWAYS** run the listed checks after backend changes.
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
  - Run frontend checks in `frontend`: `pnpm check`, `pnpm lint`, `pnpm format`
  - Run backend checks in `backend-go`: `go fmt ./...`, `go vet ./...`
  - **ALWAYS** run these after you are done making changes
- Rules / conventions:
  - **ALWAYS** use `/tmp` for any build artifacts generated by the agent.
  - **ALWAYS** treat `Taskfile.yml` and `claude-code-github-bot-*.sh` scripts as references unless explicitly asked to edit them.
  - **NEVER** invoke `task watch-frontend`, `task watch-backend`, or any command that starts a dev server unless explicitly asked.
  - **NEVER** search for or execute additional test suites unless explicitly requested.
  - **ALWAYS** run the listed checks after changes.
- Useful files:
  - `Taskfile.yml`
  - `claude-code-github-bot-setup.sh`
  - `claude-code-github-bot-cleanup.sh`

## Key Architectural Patterns
- **Schema + Persistence:** GORM models (`backend-go/models`) are the authoritative schema, applied through AutoMigrate at startup.
- **HTTP Shape:** Backend routes are grouped under `/api`; list endpoints return domain data with pagination metadata, and failures return `{ "error": string }`.
- **Rate Limiting:** Global Fiber limiter wraps all requests, with tighter per-route limiters on sensitive POST endpoints in `backend-go/routes/routes.go`.
- **Worker Orchestration:** Register workers via `WorkerManager`, persist worker state in DB, and avoid ad-hoc goroutines outside established worker flows.
- **Ranking + Stats:** Rank and statistics calculations are centralized in `backend-go/utils/utils.go` and worker implementations, not in handlers.
- **Frontend Data Flow:** Route loaders (`+page.ts`) fetch initial page data using `PUBLIC_API_URL`; avoid private env access from client-side code.
- **UI Conventions:** Reuse shadcn-svelte primitives and local lint constraints (`local/no-nested-interactive`) rather than introducing custom interaction patterns.
