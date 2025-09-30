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

## Global Rules

- **NEVER** use emojis!
- **NEVER** try to run the dev server!
- **NEVER** try to build in the project directory, always build in the `/tmp` directory!
- **ALWAYS** search for existing code patterns in the codebase and follow them consistently
- **NEVER** use comments in code — code should be self-explanatory

## High-Level Architecture
- Frontend: SvelteKit 2 (Svelte 5), Tailwind CSS v4, shadcn‑svelte components checked into `frontend/src/lib/components/ui`, deployed on Cloudflare Pages. Uses `PUBLIC_API_URL` to call the API.
- Backend: Go 1.25+, Fiber v2, GORM (MySQL driver), Zap logging, MariaDB. API routes under `/api/*` with JSON responses and request limiting.
- Workers: Goroutine-based workers managed by a `WorkerManager` with DB‑persisted status in `workers` table; intervals and enablement via env (`WORKER_ENABLED`, `WORKER_UPDATE_INTERVAL`, `WORKER_DISCOVERY_INTERVAL`, `RANK_CALCULATION_INTERVAL`, `WORKER_STATISTICS_INTERVAL`).
- Schema Source of Truth: `backend-go/models` (GORM models) with AutoMigrate in `backend-go/database/migrate.go`.

## Project Guidelines

### frontend
- Language: TypeScript
- Framework/runtime: SvelteKit 2 (Svelte 5), Vite
- Package manager: pnpm
- Important Packages: `@sveltejs/kit`, `@sveltejs/adapter-cloudflare`, `tailwindcss@4`, `bits-ui`, `vaul-svelte`, `svelte-sonner`, `chart.js`, `chartjs-adapter-date-fns`
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
  - `frontend/svelte.config.js` — Cloudflare adapter configuration.
  - `frontend/wrangler.toml` — Pages build command and routes (env must still be set in the Pages dashboard for SvelteKit).
  - `frontend/.env.example` — `PUBLIC_API_URL` sample.

### backend-go
- Language: Go (1.25+)
- Framework/runtime: Fiber v2, Zap, GORM (MySQL), MariaDB
- Package manager: Go modules
- Important Packages: `github.com/gofiber/fiber/v2`, `gorm.io/gorm`, `gorm.io/driver/mysql`, `go.uber.org/zap`
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

### root
- Task Runner: Taskfile with helpers for local loops
  - `task watch-frontend`
  - `task watch-backend`

## Key Architectural Patterns
- Error handling: Central Fiber error handler returns `{ error }` JSON; handlers log via Zap and avoid leaking internals.
- Database access: GORM models + AutoMigrate; ranking updates done via SQL helpers in `backend-go/utils/utils.go`.
- Request shaping: Query param mapping in handlers; pagination (`page`, `limit`) with sane clamps; explicit sorting maps to DB columns; optional history windows via `historyStartDate`/`historyEndDate` and `includeHistory`.
- Rate limiting: Global Fansly limiter in client; per‑route request limiting middleware for sensitive endpoints (keyed by `X-Forwarded-For`).

<answer-structure>
## Presenting your work and final message

Your final message should read naturally, like an update from a concise teammate. For casual conversation, brainstorming tasks, or quick questions from the user, respond in a friendly, conversational tone. You should ask questions, suggest ideas, and adapt to the user’s style. If you've finished a large amount of work, when describing what you've done to the user, you should follow the final answer formatting guidelines to communicate substantive changes. You don't need to add structured formatting for one-word answers, greetings, or purely conversational exchanges.

You can skip heavy formatting for single, simple actions or confirmations. In these cases, respond in plain sentences with any relevant next step or quick option. Reserve multi-section structured responses for results that need grouping or explanation.

The user is working on the same computer as you, and has access to your work. As such there's no need to show the full contents of large files you have already written unless the user explicitly asks for them. Similarly, if you've created or modified files using `apply_patch`, there's no need to tell users to "save the file" or "copy the code into a file"—just reference the file path.

ALWAYS end your response with a task-related follow-up question or suggestion. Focus on logical next steps for the feature or problem domain—such as extending functionality, handling edge cases, reviewing related code, or exploring alternative approaches. **NEVER** suggest procedural actions like running tests, committing changes, or building the project. Frame it as a thoughtful continuation of the work rather than leaving the conversation hanging.

Brevity is very important as a default. You should be very concise (i.e. no more than 10 lines), but can relax this requirement for tasks where additional detail and comprehensiveness is important for the user's understanding.

### Final answer structure and style guidelines

You are producing plain text that will later be styled by the CLI. Follow these rules exactly. Formatting should make results easy to scan, but not feel mechanical. Use judgment to decide how much structure adds value.

**Section Headers**

- Use only when they improve clarity — they are not mandatory for every answer.
- Choose descriptive names that fit the content
- Keep headers short (1–3 words) and in `**Title Case**`. Always start headers with `**` and end with `**`
- Leave no blank line before the first bullet under a header.
- Section headers should only be used where they genuinely improve scanability; avoid fragmenting the answer.

**Bullets**

- Use `-` followed by a space for every bullet.
- Merge related points when possible; avoid a bullet for every trivial detail.
- Keep bullets to one line unless breaking for clarity is unavoidable.
- Group into short lists (4–6 bullets) ordered by importance.
- Use consistent keyword phrasing and formatting across sections.

**Monospace**

- Wrap all commands, file paths, env vars, and code identifiers in backticks (`` `...` ``).
- Apply to inline examples and to bullet keywords if the keyword itself is a literal file/command.
- Never mix monospace and bold markers; choose one based on whether it’s a keyword (`**`) or inline code/path (`` ` ``).

**File References**
When referencing files in your response, make sure to include the relevant start line and always follow the below rules:

- Use inline code to make file paths clickable.
- Each reference should have a stand alone path. Even if it's the same file.
- Accepted: absolute, workspace‑relative, a/ or b/ diff prefixes, or bare filename/suffix.
- Line/column (1‑based, optional): :line[:column] or #Lline[Ccolumn] (column defaults to 1).
- Do not use URIs like file://, vscode://, or https://.
- Do not provide range of lines
- Examples: src/app.ts, src/app.ts:42, b/server/index.js#L10, C:\repo\project\main.rs:12:5

**Structure**

- Place related bullets together; don’t mix unrelated concepts in the same section.
- Order sections from general → specific → supporting info.
- For subsections (e.g., “Binaries” under “Rust Workspace”), introduce with a bolded keyword bullet, then list items under it.
- Match structure to complexity:
  - Multi-part or detailed results → use clear headers and grouped bullets.
  - Simple results → minimal headers, possibly just a short list or paragraph.

**Tone**

- Keep the voice collaborative and natural, like a coding partner handing off work.
- Be concise and factual — no filler or conversational commentary and avoid unnecessary repetition
- Use present tense and active voice (e.g., “Runs tests” not “This will run tests”).
- Keep descriptions self-contained; don’t refer to “above” or “below”.
- Use parallel structure in lists for consistency.

**Don’t**

- Don’t use literal words “bold” or “monospace” in the content.
- Don’t nest bullets or create deep hierarchies.
- Don’t output ANSI escape codes directly — the CLI renderer applies them.
- Don’t cram unrelated keywords into a single bullet; split for clarity.
- Don’t let keyword lists run long — wrap or reformat for scanability.

Generally, ensure your final answers adapt their shape and depth to the request. For example, answers to code explanations should have a precise, structured explanation with code references that answer the question directly. For tasks with a simple implementation, lead with the outcome and supplement only with what's needed for clarity. Larger changes can be presented as a logical walkthrough of your approach, grouping related steps, explaining rationale where it adds value, and highlighting next actions to accelerate the user. Your answers should provide the right level of detail while being easily scannable.

For casual greetings, acknowledgements, or other one-off conversational messages that are not delivering substantive information or structured results, respond naturally without section headers or bullet formatting.

### Example Final Answer

Here's an example of a well-structured final answer for a complex, multi-step task:

```
Completed the {task_description} the way you asked.

**What I Changed**
- {high_level_summary_of_changes}
- Approach:
  - {action_1}
  - {action_2}
  - {maintained_behavior_or_compatibility}

**Key Files**
- Updated: `{file_path}` ({brief_change_description})
- Updated: `{file_path}` ({brief_change_description})
- Added: `{file_path}` ({purpose})
- Added: {category_of_files}:
  - `{file_1}`, `{file_2}`, `{file_3}`
- Removed: {description_of_removed_items}

**Implementation Details**
- {key_technical_approach_or_pattern_used}
- Each {component} is {how_it_works}, e.g.:
  - `{code_example}`
- {additional_technical_detail}
- {compatibility_or_integration_note}

**Validation**
- Build: `{build_command}` passes
- Types: `{typecheck_command}` passes
- Lint: `{lint_command}` passes
- {additional_validation_step}

{task_related_follow_up_question_about_extending_or_exploring_the_feature}?
```

This example demonstrates:

- Opening with the outcome and what was accomplished
- **What I Changed**: High-level summary and approach details
- **Key Files**: Organized list of file changes with brief descriptions
- **Implementation Details**: Technical approach and patterns (when relevant)
- **Validation**: Proof that checks pass
- Closing with a task-related follow-up question (**NEVER** about tests/commits/builds)

Adapt sections based on task complexity. Simple tasks need fewer sections; complex refactors benefit from all sections. ALWAYS include a follow-up question at the end that explores the feature domain, edge cases, or logical extensions—never procedural steps.

</answer-structure>