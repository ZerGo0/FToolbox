# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Fanslytagstats is a full-stack web application with a SvelteKit frontend and Bun backend, designed to provide statistics and analytics for Fansly tags.

## Development Commands

### Frontend (SvelteKit)

**ALWAYS** run `pnpm check && pnpm lint` after making changes

### Backend (Bun)

**ALWAYS** run `bun check && bun lint` after making changes

## Database Migrations

**IMPORTANT**: Always use Drizzle to generate migrations. NEVER write manual SQL migration files.

Use the following command to generate migrations after schema changes:

```bash
bun drizzle-kit generate
```

## Architecture

### Tech Stack

- **Frontend**: SvelteKit with Svelte 5, TypeScript, Tailwind CSS v4, shadcn-svelte components
- **Backend**: Bun runtime with TypeScript
- **Deployment**: Frontend configured for Cloudflare adapter
- **Task Runner**: Taskfile for orchestrating development tasks

### Directory Structure

- `/frontend/` - SvelteKit application
  - `/src/lib/components/ui/` - shadcn-svelte UI components (40+ components)
  - `/src/routes/` - SvelteKit file-based routing
  - Uses Svelte 5 features including `.svelte.ts` files
- `/backend/` - Bun API server (currently minimal placeholder)
- `/site-frontend/` - Empty directory (possibly for landing page)

### Key Configuration

- Frontend uses `pnpm` with specific build-only dependencies for performance
- Backend uses `bun` for both runtime and package management
- Taskfile automatically kills existing processes on ports before starting new ones
- Frontend configured with Cloudflare adapter for deployment

### Current State

The project is in initial setup phase with frontend UI components installed but business logic not yet implemented. The backend is a placeholder server returning "Welcome to Bun!" on port 3000.
