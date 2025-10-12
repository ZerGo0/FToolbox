# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
- Name: FToolbox
- Description: Full-stack web application delivering analytics and tooling for Fansly creators and consumers with a SvelteKit frontend and Go backend.
- Key notes/warnings:
  - Database schema lives in `backend-go/models` and AutoMigrate applies changes on startup; update models instead of issuing manual migrations.
  - Server listens on `PORT` (default `3000` from `backend-go/config/config.go`); the Taskfile kills `3001` solely for the Air loop—bind to `PORT` for anything hitting the API.
  - Tag heat is deprecated and responses force `heat = 0`; never add UI or metrics that depend on legacy heat values.
  - Fansly API calls must go through `backend-go/fansly/client.go`, which enforces global rate limiting and 429 retry logic configured by `FANSLY_GLOBAL_RATE_LIMIT` and `FANSLY_GLOBAL_RATE_LIMIT_WINDOW`.
  - Cloudflare Pages builds require frontend env vars (e.g. `PUBLIC_API_URL`) set in the Pages dashboard; the `wrangler.toml [vars]` block is ignored during SvelteKit builds.

## Global Rules
- **NEVER** use emojis!
- **NEVER** try to run the dev server!
- **NEVER** try to build in the project directory, always build in the `/tmp` directory!
- **ALWAYS** search for existing code patterns in the codebase and follow them consistently
- **NEVER** use comments in code - code should be self-explanatory
- **NEVER** cut corners, don't leave comments like `TODO: Implement X in the future here`! Always fully implement everything!
- **ALWAYS** when you're done, self-critique your work until you're sure it's correct 
- **NEVER** trigger automated test suites unless a human explicitly asks; stick to the documented checks only.

## High-Level Architecture
- **Frontend:** SvelteKit 2 (Svelte 5) with Tailwind CSS v4 and shadcn-svelte UI, deployed on Cloudflare Pages; fetches the Go API using absolute URLs derived from `PUBLIC_API_URL`.
- **Backend:** Go 1.25 Fiber API with Zap logging, GORM (MySQL driver) targeting MariaDB, global middleware for compression, CORS, ETag, and request limiting, serving JSON responses under `/api/*`.
- **Workers:** Goroutine workers managed by `backend-go/workers/manager.go`; enablement and intervals come from env vars and statuses persist in the `workers` table for visibility.
- **Fansly Integration:** `backend-go/fansly/client.go` centralizes authentication, retries, and global rate limiting for Fansly requests; all scrapers and handlers rely on this client.
- **Schema Source of Truth:** `backend-go/models` drives AutoMigrate via `backend-go/database/migrate.go`; never mutate the database outside this layer.
- **Infrastructure Utilities:** `Taskfile.yml` scripts orchestrate human dev loops by killing ports and starting watchers—treat them as references only; do not invoke them here.

## Project Guidelines

### frontend
- Language: TypeScript
- Framework/runtime: SvelteKit 2 (Svelte 5), Vite
- Package manager: pnpm
- Important Packages: `@sveltejs/kit`, `@sveltejs/adapter-cloudflare`, `tailwindcss@4`, `bits-ui`, `svelte-sonner`, `chart.js`, `chartjs-adapter-date-fns`
- Checks:
  - Syntax Check: `pnpm check`
  - Lint: `pnpm lint`
  - Format: `pnpm format`
  - **ALWAYS** run these after you are done making changes
- Rules / conventions:
  - **ALWAYS** build API requests with `$env/static/public` values (notably `PUBLIC_API_URL`) and send absolute URLs from loaders or server-only modules.
  - **ALWAYS** compose UI from shadcn-svelte primitives in `frontend/src/lib/components/ui` before rolling new components.
  - **NEVER** nest interactive children inside trigger components; follow the `local/no-nested-interactive` ESLint rule and reuse the helper variants.
  - **NEVER** start dev loops, previews, or automated tests; limit execution to the listed checks unless a human explicitly requests otherwise.
  - **ALWAYS** sync Cloudflare Pages deployment metadata with `wrangler.toml` and the Pages dashboard env configuration.
- Useful files:
  - `frontend/eslint.config.js` — Central Svelte/TypeScript lint configuration and plugin wiring.
  - `frontend/eslint-plugin-local/index.js` — Implements `no-nested-interactive`.
  - `frontend/svelte.config.js` — Cloudflare adapter setup.
  - `frontend/wrangler.toml` — Pages deployment commands and bindings.
  - `frontend/.env.example` — Sample `PUBLIC_API_URL`.

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
  - **ALWAYS** load settings through `config.Load()` so default intervals and rate limits apply, and pass config explicitly.
  - **ALWAYS** persist schema changes by editing `backend-go/models`; rely on AutoMigrate for DDL.
  - **ALWAYS** return `{ "error": string }` JSON and log diagnostics with `zap.L()` instead of exposing internals.
  - **NEVER** bypass the shared `fansly.Client` or global limiters when calling external APIs.
  - **NEVER** spin up worker loops, Air, or other dev tooling from the agent; limit changes to code paths.
  - **NEVER** run automated tests beyond `go fmt`/`go vet` unless a human explicitly asks.
- Useful files:
  - `backend-go/config/config.go` — Environment defaults for DB, server, workers, and rate limiting.
  - `backend-go/database/connection.go` / `backend-go/database/migrate.go` — Connection lifecycle and AutoMigrate wiring.
  - `backend-go/routes/routes.go` — API routing and per-route request limiters.
  - `backend-go/fansly/client.go` — Fansly HTTP client with shared limiter.
  - `backend-go/workers/*` — Worker definitions managed by `WorkerManager`.
  - `backend-go/utils/utils.go` — Rank calculation utilities.

### root
- Language: Mixed (YAML, shell)
- Framework/runtime: Taskfile helpers
- Package manager: pnpm (frontend), Go modules (backend)
- Important Packages: `task`, `pnpm`, `air`
- Checks:
  - None; root automation targets human workflows only.
  - **ALWAYS** run these after you are done making changes
- Rules / conventions:
  - **NEVER** invoke `task watch-frontend` or `task watch-backend`; they start dev servers prohibited by global rules.
  - **ALWAYS** use `/tmp` for any build artifacts generated by the agent.
  - **NEVER** alter automation scripts without matching existing patterns.
- Useful files:
  - `Taskfile.yml` — Documents human dev tasks and port cleanup expectations.
  - `claude-code-github-bot-setup.sh` / `claude-code-github-bot-cleanup.sh` — Environment orchestration scripts.

## Patterns Directory
- `.zergo0/patterns/` is the canonical catalog of reusable implementation patterns generated via the `patterns` command. Consult relevant entries before building new features to stay aligned with accepted structures.
- When a suitable pattern is missing, extend the catalog under `.zergo0/patterns/` only after validating the approach against existing modules, and keep contributions focused so future automation can leverage them safely.

## Key Architectural Patterns
- **Error Handling:** Central Fiber error handler and route handlers return `{ "error": string }` JSON while logging with Zap; replicate this contract in new endpoints.
- **Request Shaping:** Pagination, sorting, and history window parsing in `backend-go/handlers/tag_handler.go` sets expectations for query handling—reuse those clamps and column maps for additional list endpoints.
- **Rate Limiting:** Global limiter middleware wraps all requests, and sensitive POST routes apply stricter per-IP limiters; wire new routes through the same limiter helpers.
- **Ranking & Statistics:** Rank calculations live in `backend-go/utils/utils.go` and run via workers; schedule new analytics through `WorkerManager` rather than ad-hoc goroutines.
- **Frontend Data Fetching:** SvelteKit loaders construct absolute API URLs using `PUBLIC_API_URL` and avoid exposing env values client-side; keep data access server-side where feasible.
- **UI Composition:** Favor shadcn-svelte components and Tailwind utility patterns from `frontend/src/lib/components/ui` to maintain consistent styling and accessibility.

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