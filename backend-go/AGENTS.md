# AGENTS.md

This file provides guidance to Codex CLI when working in this part of the repository.

## Scope
- This file governs `backend-go` and its descendants.

## Local Overview
- `backend-go` contains the Fiber API, GORM models, background workers, and the shared Fansly client.
- `models` plus `database/migrate.go` are the schema source of truth.
- `main.go` wires configuration, database connection, auto-migration, worker registration, rate limiting, and route setup.

## Local Rules
- ALWAYS load environment-backed settings through `config.Load()` and thread the resulting config through backend setup.
- ALWAYS make schema changes in `models`, and update `database/migrate.go` when a new model needs to be included in `database.AutoMigrate`.
- ALWAYS route Fansly API access through `fansly.Client` so auth, retries, and global rate limiting stay centralized.
- ALWAYS return route-level failures in `{ "error": string }` shape.
- Register recurring background work through `workers.WorkerManager`; do not introduce separate ad-hoc worker orchestration.

## Checks
- `go fmt ./...`
- `go vet ./...`
- **ALWAYS** run these after you are done making changes

## Local Patterns
- Routes are registered under `/api` in `routes/routes.go`, with tighter per-route Fiber limiters on request endpoints.
- Worker registration persists worker rows in the database and exposes status through the existing worker APIs.
- Rank and statistics computation belong in `utils` and `workers`, not in route handlers.
- Tag heat is legacy-only in the current backend; `main.go` notes that responses now force `heat = 0`.

## Useful files
- `main.go`
- `config/config.go`
- `database/connection.go`
- `database/migrate.go`
- `.env.example`
- `Dockerfile`
- `models`
- `routes/routes.go`
- `fansly/client.go`
- `workers/manager.go`
