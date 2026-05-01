<agents_md>
<purpose>
This file provides guidance to Codex CLI when working in this part of the repository.
</purpose>

<scope>
- This file governs `backend-go` and its descendants.
</scope>

<local_overview>
- `backend-go` contains the Fiber API, GORM models, background workers, and the shared Fansly client.
- `models` plus `database/migrate.go` are the schema source of truth.
- `main.go` wires configuration, database connection, auto-migration, worker registration, rate limiting, and route setup.
</local_overview>

<local_rules>
- ALWAYS load environment-backed settings through `config.Load()` and thread the resulting config through backend setup.
- ALWAYS make schema changes in `models` and rely on `database.AutoMigrate`; this repo does not maintain a separate handwritten migration layer.
- ALWAYS route Fansly API access through `fansly.Client` so auth, retries, and global rate limiting stay centralized.
- ALWAYS return route-level failures in `{ "error": string }` shape.
- Register recurring background work through `workers.WorkerManager`; do not introduce separate ad-hoc worker orchestration.
</local_rules>

<checks>
- `go fmt ./...`
- `go vet ./...`
- **ALWAYS** run these after you are done making changes
</checks>

<local_patterns>
- Routes are registered under `/api` in `routes/routes.go`, with tighter per-route Fiber limiters on request endpoints.
- Worker registration persists worker rows in the database and exposes status through the existing worker APIs.
- Rank and statistics computation belong in `utils` and `workers`, not in route handlers.
- Tag heat is legacy-only in the current backend; `main.go` notes that responses now force `heat = 0`.
</local_patterns>

<useful_files>
- `main.go`
- `config/config.go`
- `database/connection.go`
- `database/migrate.go`
- `models`
- `routes/routes.go`
- `fansly/client.go`
- `workers/manager.go`
</useful_files>
</agents_md>
