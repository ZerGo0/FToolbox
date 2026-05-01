<agents_md>
<purpose>
This file provides guidance to Codex CLI when working with code in this repository.
</purpose>

<project_overview>
- Name: FToolbox.
- Description: full-stack web application for Fansly analytics and tooling, with a SvelteKit frontend in `frontend` and a Go API in `backend-go`.
- The authoritative database schema lives in `backend-go/models`, and startup applies it through `backend-go/database/migrate.go`.
- The frontend is deployed with the SvelteKit Cloudflare adapter and depends on `PUBLIC_API_URL` at build time.
</project_overview>

<repository_wide_rules>
- Use `/tmp` for build artifacts or other generated output if a build is explicitly required.
- Do not run `task watch-frontend` or `task watch-backend` unless the user explicitly asks; both start dev servers.
- When you change a specific subproject, run only that subproject's documented static checks unless the task requires broader verification.
- **ALWAYS** at the end of your turn, ask a follow-up question for the next logical step (**DON'T** ask questions like "Should I run tests?" or "Should I lint?", only ask questions that are relevant to the task at hand)
</repository_wide_rules>

<high_level_architecture>
- `frontend`: SvelteKit 2 + Svelte 5 single-page app with the Cloudflare adapter; client code calls the backend under `/api`.
- `backend-go`: Fiber API with Zap logging and GORM against MariaDB/MySQL.
- `backend-go/workers`: background jobs registered through `WorkerManager`; worker status is stored in the `workers` table and surfaced via `/api/workers/status`.
- `backend-go/fansly`: shared Fansly client that centralizes auth, retries, and global rate limiting.
</high_level_architecture>

<repository_workflow>
- Frontend work uses `pnpm` from `frontend`.
- Backend work uses Go modules from `backend-go`.
- `task install` runs `pnpm install -r` and `go mod download` for every Go module.
- `Taskfile.yml` is mainly a human workflow entrypoint; its watch tasks start dev servers.
- There is no single repo-wide lint, typecheck, or test command in `Taskfile.yml`.
- `.github/workflows/deploy-backend-go.yml` is the current deployment workflow and builds the backend Docker image from `backend-go`.
</repository_workflow>

<project_guidelines>
- `frontend`: TypeScript SvelteKit app. See `frontend/AGENTS.md` for local commands, UI rules, and Cloudflare build-time environment guidance.
- `backend-go`: Go Fiber API. See `backend-go/AGENTS.md` for local commands, schema rules, worker rules, and backend request handling conventions.
</project_guidelines>

<key_architectural_patterns>
- Backend routes are grouped under `/api`.
- Route-level failures should surface as JSON objects with an `error` string.
- Frontend API calls are expected to target absolute URLs built from `PUBLIC_API_URL`.
- Background jobs should use the existing worker manager flow instead of introducing separate recurring job mechanisms.
</key_architectural_patterns>
</agents_md>
