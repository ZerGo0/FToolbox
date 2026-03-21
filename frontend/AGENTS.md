# AGENTS.md

This file provides guidance to Codex CLI when working in this part of the repository.

## Scope

- This file governs `frontend` and its descendants.

## Local Overview

- `frontend` is the SvelteKit 2 + Svelte 5 client application deployed with the Cloudflare adapter.
- The app currently runs client-side only via `src/routes/+layout.ts`.
- `frontend/eslint-plugin-local` supplies the custom `local/no-nested-interactive` rule that is enabled in `eslint.config.js`.

## Local Rules

- ALWAYS build backend requests from `$env/static/public` via `PUBLIC_API_URL`; do not hardcode backend hosts.
- ALWAYS reuse the existing UI primitives under `src/lib/components/ui` before introducing new component patterns.
- NEVER nest interactive controls inside trigger components such as `DialogTrigger`; use the direct child or snippet-based composition pattern expected by `local/no-nested-interactive`.
- Cloudflare Pages builds need `PUBLIC_API_URL` configured in the Pages dashboard; `wrangler.toml [vars]` is not enough for SvelteKit build-time env resolution.

## Checks

- `pnpm check`
- `pnpm lint`
- `pnpm format`
- **ALWAYS** run these after you are done making changes

## Local Patterns

- `svelte.config.js` defines the Cloudflare adapter and the `@/*` alias to `src/lib/*`.
- Route loaders and UI components fetch backend data from `${PUBLIC_API_URL}/api/...`.
- The local ESLint plugin is the source of truth for the trigger/button nesting rule and its expected composition pattern.

## Useful files

- `package.json`
- `eslint.config.js`
- `svelte.config.js`
- `wrangler.toml`
- `.env.example`
- `src/routes/+layout.ts`
- `eslint-plugin-local/index.js`
