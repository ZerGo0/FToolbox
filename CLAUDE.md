# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
- Name: FToolbox
- Description: Full-stack SvelteKit + Go application delivering Fansly analytics with a Cloudflare Pages frontend, Fiber API, and background workers backed by MariaDB.
- Key notes/warnings:
  - Database schema lives in `backend-go/models` and is migrated automatically on startup via `backend-go/database/migrate.go`; schema edits must go through the GORM models.
  - `PORT` controls the API listener (default `3000` from `backend-go/config/config.go`); the Taskfile kills `3001` for the Air loop—always respect `PORT` rather than assuming alternate ports.
  - Tag `heat` is deprecated and forced to `0`; never build UI logic around it.
  - Fansly API access uses the shared rate limiter in `backend-go/fansly/client.go`; configure `FANSLY_GLOBAL_RATE_LIMIT`, `FANSLY_GLOBAL_RATE_LIMIT_WINDOW`, and optional `FANSLY_AUTH_TOKEN` before making new calls.
  - Cloudflare Pages builds need frontend env vars (e.g. `PUBLIC_API_URL`) configured in the Pages dashboard; SvelteKit ignores `wrangler.toml [vars]` during build.
  - Background workers persist state in the `workers` table and are orchestrated by `WorkerManager`; honor interval env vars (`WORKER_ENABLED`, `WORKER_UPDATE_INTERVAL`, `WORKER_DISCOVERY_INTERVAL`, `RANK_CALCULATION_INTERVAL`, `WORKER_STATISTICS_INTERVAL`).
  - Do not run dev servers or watchers; the Taskfile commands start servers that violate the global rules.
  - Never run automated test suites or hunt for them unless a human explicitly requests it; rely on static analysis commands only.

## Global Rules

- **NEVER** use emojis!
- **NEVER** try to run the dev server!
- **NEVER** try to build in the project directory, always build in the `/tmp` directory!
- **ALWAYS** search for existing code patterns in the codebase and follow them consistently
- **NEVER** use comments in code - code should be self-explanatory
- **NEVER** cut corners, don't leave comments like `TODO: Implement X in the future here`! Always fully implement everything!
- **ALWAYS** when you're done, self-critique your work until you're sure it's correct

## High-Level Architecture
- **Frontend:** SvelteKit 2 (Svelte 5) served via Cloudflare Pages, Tailwind CSS v4, shadcn-svelte UI components in `frontend/src/lib/components/ui`, fetching data from the Go API using absolute URLs built with `PUBLIC_API_URL`.
- **Backend API:** Go 1.25 Fiber service with Zap logging, GORM (MySQL driver) against MariaDB, exposing `/api/*` routes with request limiting and JSON responses.
- **Background Workers:** Goroutine workers managed by `backend-go/workers/manager.go` polling Fansly and updating creators/tags on configured intervals, persisting status in the `workers` table.
- **Fansly Integration:** `backend-go/fansly/client.go` centralizes HTTP access, retries, and global rate limiting; `GetSuggestionsData` powers both tag and creator flows.
- **Schema Source of Truth:** `backend-go/models` GORM structs plus AutoMigrate in `backend-go/database/migrate.go` govern database structure.

## Project Guidelines

### frontend
- Language: TypeScript
- Framework/runtime: SvelteKit 2 (Svelte 5) on Vite
- Package manager: pnpm
- Important Packages: `@sveltejs/kit`, `@sveltejs/adapter-cloudflare`, `tailwindcss@4`, `bits-ui`, `vaul-svelte`, `svelte-sonner`, `chart.js`, `chartjs-adapter-date-fns`, `sveltekit-superforms`
- Checks:
  - Syntax Check: `pnpm check`
  - Lint: `pnpm lint`
  - Format: `pnpm format`
  - **ALWAYS** run these after you are done making changes
- Rules/conventions:
  - **ALWAYS** build absolute fetch URLs using `$env/static/public` `PUBLIC_API_URL`.
  - **ALWAYS** use shadcn-svelte components from `frontend/src/lib/components/ui` and existing utility helpers (`frontend/src/lib/utils.ts`) before adding new ones.
  - **NEVER** nest interactive children inside Trigger components; follow the ESLint `local/no-nested-interactive` guidance (see existing implementations for the child-snippet pattern).
  - **NEVER** start `pnpm dev`, `pnpm preview`, or Taskfile watchers; rely on static generation and analysis only.
  - **NEVER** add code comments; keep Svelte components self-explanatory via clear structure and naming.
- Useful files:
  - `frontend/eslint.config.js` – project ESLint setup with the custom rule.
  - `frontend/eslint-plugin-local/index.js` – definition of `no-nested-interactive`.
  - `frontend/svelte.config.js` – Cloudflare adapter configuration.
  - `frontend/wrangler.toml` – Pages build/deploy commands (env vars still come from the dashboard).
  - `frontend/.env.example` – sample `PUBLIC_API_URL`.
  - `frontend/src/routes/creators/+page.svelte` & `+page.ts` – canonical data loading + UI pattern for statistics pages.

### backend-go
- Language: Go (1.25+)
- Framework/runtime: Fiber v2 with Zap logging
- Package manager: Go modules
- Important Packages: `github.com/gofiber/fiber/v2`, `gorm.io/gorm`, `gorm.io/driver/mysql`, `go.uber.org/zap`, `github.com/joho/godotenv`
- Checks:
  - Format: `go fmt ./...`
  - Vet: `go vet ./...`
  - **ALWAYS** run these after you are done making changes
- Rules/conventions:
  - **ALWAYS** route errors through the centralized Fiber error handler and respond with `{ "error": string }` without leaking internal messages.
  - **ALWAYS** obtain the global logger via `zap.L()` and ensure structured fields for context.
  - **ALWAYS** modify database structure via `backend-go/models/*` so AutoMigrate keeps the schema authoritative.
  - **ALWAYS** reuse the Fansly client helpers (`GetSuggestionsData`, rate limiter) for new API interactions to avoid duplicating HTTP logic.
  - **NEVER** run `air` or other live-reload servers; keep changes static.
  - **NEVER** assume the deprecated `heat` value conveys meaning; it is hardcoded to `0`.
- Useful files:
  - `backend-go/config/config.go` – environment defaults and parsing helpers.
  - `backend-go/database/connection.go` – DB connection setup.
  - `backend-go/database/migrate.go` – AutoMigrate entry point.
  - `backend-go/routes/routes.go` – complete routing table with per-route limiters.
  - `backend-go/fansly/client.go` – HTTP client with limiter/backoff.
  - `backend-go/workers/*` – worker manager plus tag/creator updater implementations.
  - `backend-go/models/*` – schema definitions for tags, creators, history, statistics, and workers.

### root
- Language: Mixed (Go + TypeScript projects)
- Framework/runtime: Taskfile-based command wrappers
- Package manager: pnpm (frontend), Go modules (backend)
- Important Packages: N/A (use project-level dependencies)
- Checks:
  - Delegate to project-specific commands above
  - **ALWAYS** run these after you are done making changes
- Rules/conventions:
  - **NEVER** invoke `task watch-frontend` or `task watch-backend`; they start dev servers on ports `5173` and `3001`, which violates the global rules.
  - **ALWAYS** review `current_task.md` for recent context before extending functionality to stay aligned with in-progress initiatives.
  - **NEVER** add repository-wide comments or placeholder directives; finish implementations fully.
  - **NEVER** run automated test suites unless explicitly directed by a human.
- Useful files:
  - `Taskfile.yml:1` – task definitions (for reference only; do not execute server tasks).
  - `current_task.md:1` – snapshot of the most recent feature work and expectations.

## Patterns Directory
- `.zergo0/patterns/` is the canonical catalog produced by the `patterns` command; consult it before introducing new patterns so automated agents can stay consistent. If the directory is missing, generate it with the patterns tooling before relying on it.
- Review the relevant pattern doc whenever you touch a feature area, mirror the described structure, and only add new pattern files when a repeatable approach has stabilized. Keep any additions concise, action-oriented, and immediately useful for future automation.
- When extending the catalog, ensure examples compile, reference actual repository code, and avoid speculative guidance so downstream agents can trust the automation scripts that consume these patterns.

## Key Architectural Patterns
- **Error handling:** Central Fiber error middleware returns `{ "error": string }` JSON; handlers log via `zap.L()` with contextual fields and avoid leaking implementation details.
- **Database access:** All persistence goes through GORM models in `backend-go/models`; AutoMigrate applies schema updates, and workers use transactions to keep tag/creator history consistent.
- **Fansly client usage:** `backend-go/fansly/client.go` encapsulates retries, backoff, and the global limiter—reuse `GetSuggestionsData` for suggestions-derived data and respect rate-limit configuration.
- **Workers:** `WorkerManager` coordinates long-running goroutines with DB-backed status; intervals are environment-driven, and creator/tag updates reuse shared helper methods to minimize API calls.
- **Frontend data flow:** SvelteKit load functions (e.g., `frontend/src/routes/creators/+page.ts`) fetch from the API with absolute URLs and pass structured data to UI components; charts use `chart.js` with date-fns adapters for history visualization.
- **UI composition:** shadcn-svelte primitives live in `frontend/src/lib/components/ui`; combine them with Tailwind utility classes and helper utilities instead of introducing bespoke styling systems.
- **Testing posture:** Rely on lint, format, and type-check commands for validation; never trigger test suites or search for them unless a human explicitly requests execution.

<answer-structure>
## MANDATORY Answer Format

**CRITICAL:** You MUST follow these exact formatting rules for ALL responses. No exceptions.

**ABSOLUTE Requirements:**
- ALWAYS use the exact structure below
- NEVER deviate from the specified format
- ALWAYS end with a task-related follow-up question
- ALWAYS keep responses concise (≤10 lines unless technical details require more)

**EXACT Structure Template:**
```
Completed {task}.

**What I Changed**
- {high_level_summary}
- {key_approach_points}

**Key Files**
- Updated: `{file_path}` ({brief_change})
- Added: `{file_path}` ({purpose})

**Implementation Details**
- {technical_approach}
- {key_code_example}

**Validation**
- Types: `{typecheck_command}` passes
- Lint: `{lint_command}` passes
- {additional_validation_step}

{task_related_follow_up_question}
```

**NON-NEGOTIABLE Formatting Rules:**
- Headers: EXACTLY `**Title Case**` (1-3 words only)
- Bullets: ALWAYS start with `- ` (dash, space)
- Monospace: ALWAYS use backticks for commands, paths, code
- File references: ALWAYS use `path:line` format
- NEVER use conversational tone outside the follow-up question
- NEVER mention saving files or copying code
- NEVER suggest procedural actions (tests, commits, builds)

**MANDATORY Follow-up Questions:**
- MUST relate to extending functionality
- MUST explore edge cases or alternatives
- MUST be task-related (never procedural)

**NO EXCEPTIONS** to these rules regardless of request type, complexity, or perceived user intent.
</answer-structure>