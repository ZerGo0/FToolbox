# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
- Name: FToolbox
- Description: Full-stack web application providing tools and insights for Fansly creators and consumers with a SvelteKit frontend and a Go backend backed by MariaDB.
- Key notes/warnings:
  - Database schema is defined by GORM models in `backend-go/models` and applied via AutoMigrate on startup; schema changes belong in those models.
  - Backend listens on `PORT` (defaults to `3000` in `backend-go/config/config.go`). The Taskfile only clears `3001` for the Air loop; depend on `PORT` for server configuration.
  - Tag `heat` is deprecated and forced to `0` in responses; never build UI logic that depends on it.
  - Fansly API access uses a global rate limiter with retries around HTTP 429; configure via `FANSLY_GLOBAL_RATE_LIMIT`, `FANSLY_GLOBAL_RATE_LIMIT_WINDOW`, and optional `FANSLY_AUTH_TOKEN`.
  - Frontend environment variables such as `PUBLIC_API_URL` must be set in the Cloudflare Pages dashboard so SvelteKit sees them at build time; `wrangler.toml` vars are not injected automatically.
  - Taskfile watchers terminate ports `5173`/`3001` before and after local loops; keep them for human workflows but do not invoke them here because they start dev servers.

## Global Rules

- **NEVER** use emojis!
- **NEVER** try to run the dev server!
- **NEVER** try to build in the project directory, always build in the `/tmp` directory!
- **ALWAYS** search for existing code patterns in the codebase and follow them consistently
- **NEVER** use comments in code - code should be self-explanatory
- **NEVER** cut corners, don't leave comments like `TODO: Implement X in the future here`! Always fully implement everything!
- **ALWAYS** when you're done, self-critique your work until you're sure it's correct

## High-Level Architecture
- Frontend: SvelteKit 2 (Svelte 5) with Vite, Tailwind CSS v4, and shadcn-svelte components stored in `frontend/src/lib/components/ui`; deployed to Cloudflare Pages and reaches the API via `PUBLIC_API_URL`.
- Backend: Go 1.25+ service built on Fiber v2, Zap logging, and GORM with the MySQL driver; exposes JSON endpoints under `/api/*` with request limiting and centralized error handling.
- Workers: Goroutine-based workers managed by a `WorkerManager` persisted in the `workers` table; intervals controlled via env (`WORKER_ENABLED`, `WORKER_UPDATE_INTERVAL`, `WORKER_DISCOVERY_INTERVAL`, `RANK_CALCULATION_INTERVAL`, `WORKER_STATISTICS_INTERVAL`).
- Database: MariaDB accessed exclusively through GORM; migrations run via AutoMigrate during startup.
- Schema Source of Truth: `backend-go/models` combined with `backend-go/database/migrate.go`.
- Observability: Zap global logger (`zap.L()`) handles structured logging across handlers and workers.

## Project Guidelines

### frontend
- Language: TypeScript
- Framework/runtime: SvelteKit 2 (Svelte 5) with Vite and the Cloudflare adapter
- Package manager: pnpm
- Important Packages: `@sveltejs/kit`, `@sveltejs/adapter-cloudflare`, `tailwindcss@4`, `bits-ui`, `vaul-svelte`, `svelte-sonner`, `chart.js`, `chartjs-adapter-date-fns`, `layerchart`
- Checks:
  - Syntax Check: `pnpm check`
  - Lint: `pnpm lint`
  - Format: `pnpm format`
  - **ALWAYS** run these after you are done making changes
- Rules/conventions:
  - **ALWAYS** pull UI primitives from `frontend/src/lib/components/ui` (shadcn-svelte) before introducing new widgets.
  - **NEVER** nest interactive components inside Trigger components; rely on the child-snippet pattern or `class={buttonVariants(...)}` on the trigger.
  - **ALWAYS** read API endpoints from `$env/static/public` (`PUBLIC_API_URL`) and construct absolute URLs.
  - **NEVER** assume Cloudflare Pages injects env vars automatically; declare required build-time vars explicitly.
- Useful files:
  - `frontend/eslint.config.js`
  - `frontend/eslint-plugin-local/index.js`
  - `frontend/svelte.config.js`
  - `frontend/wrangler.toml`
  - `frontend/.env.example`

### backend-go
- Language: Go 1.25+
- Framework/runtime: Fiber v2 with Zap logging and GORM (MySQL driver) backed by MariaDB
- Package manager: Go modules
- Important Packages: `github.com/gofiber/fiber/v2`, `gorm.io/gorm`, `gorm.io/driver/mysql`, `go.uber.org/zap`, `github.com/joho/godotenv`
- Checks:
  - Format: `go fmt ./...`
  - Vet: `go vet ./...`
  - **ALWAYS** run these after you are done making changes
- Rules/conventions:
  - **ALWAYS** use the global Zap logger (`zap.L()`) and return JSON errors as `{ "error": string }` without leaking internals.
  - **ALWAYS** evolve schema by editing models under `backend-go/models`; let AutoMigrate apply migrations.
  - **NEVER** bypass the Fansly global rate limiter; reuse the client in `backend-go/fansly/client.go`.
  - **NEVER** rely on the deprecated `heat` tag; data is forced to `0`.
- Useful files:
  - `backend-go/config/config.go`
  - `backend-go/database/connection.go`
  - `backend-go/database/migrate.go`
  - `backend-go/fansly/client.go`
  - `backend-go/routes/routes.go`
  - `backend-go/handlers/*`
  - `backend-go/workers/*`
  - `backend-go/.env.example`

### root
- Language: YAML and shell for automation glue
- Framework/runtime: Task (go-task) for local loops orchestrating frontend/backend tooling
- Package manager: pnpm (frontend) and Go modules (backend), managed per project
- Important Packages: Taskfile runner, pnpm, Air (backend live reload)
- Checks:
  - Syntax Check: run project-specific commands above as needed
  - Lint: run project-specific commands above as needed
  - Format: run project-specific commands above as needed
  - **ALWAYS** run these after you are done making changes
- Rules/conventions:
  - **NEVER** execute Taskfile watch targets or any dev server locally; they are documented for human workflows only.
  - **ALWAYS** perform builds or heavier tooling in `/tmp`, never within the repo tree.
  - **NEVER** run or recommend automated test suites unless a human explicitly asks; do not search for test commands proactively.
- Useful files:
  - `Taskfile.yml`
  - `current_task.md`
  - `claude-code-github-bot-setup.sh`
  - `claude-code-github-bot-cleanup.sh`

## Patterns Directory
- `.zergo0/patterns/` is the canonical catalog that the `patterns` command generates to codify implementation patterns for this repository. Review it before starting new features to reuse established approaches and keep automation aligned.
- When extending the catalog, add narrowly scoped entries that reflect code already merged, keep naming consistent with existing documents, and ensure examples stay up to date. If the directory is absent, create it alongside the first pattern export so future automation can rely on a stable location.
- Use these pattern documents during planning to decide whether new code should follow an existing blueprint or whether a carefully justified extension is required.

## Key Architectural Patterns
- **Error handling:** Central Fiber error handler returns `{ "error": string }`; handlers log via `zap.L()` and never leak internal errors.
- **Database access:** All tables map to structs in `backend-go/models`; AutoMigrate enforces schema, and ranking utilities live in `backend-go/utils/utils.go`.
- **Request shaping:** Handlers clamp pagination (`page`, `limit`), map query params to whitelisted columns, and support optional history windows via `historyStartDate`, `historyEndDate`, and `includeHistory`.
- **Rate limiting:** Fansly client enforces global throttling with retry/backoff around HTTP 429 responses; per-route middleware adds request limits keyed by `X-Forwarded-For`.
- **Frontend data flow:** SvelteKit load functions call the API through absolute URLs built with `PUBLIC_API_URL`, and shared UI comes from shadcn-svelte components.
- **Local execution discipline:** Favor static analysis commands listed above; never trigger test suites or long-running dev services without explicit human direction.

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