# AGENTS.md

This file provides guidance to Codex CLI when working in this part of the repository.

## Scope
- This file governs `frontend` and its descendants.

## Local Overview
- `frontend` is the SvelteKit 2 + Svelte 5 client application.
- The app currently runs client-side only via `src/routes/+layout.ts`.
- The checked-in deployment path is `wrangler.toml` plus `pnpm deploy:production`; `README.md` still mentions Cloudflare Pages, so prefer the config and scripts when they disagree.
- `frontend/eslint-plugin-local` supplies the custom `local/no-nested-interactive` rule that is enabled in `eslint.config.js`.

## Local Rules

- ALWAYS build backend requests from `$env/static/public` via `PUBLIC_API_URL`; do not copy older hardcoded-host calls such as the localhost request that still exists in `src/lib/components/PostTextDialog.svelte`.
- ALWAYS reuse the existing UI primitives under `src/lib/components/ui` before introducing new component patterns.
- NEVER nest interactive controls inside trigger components such as `DialogTrigger`; use the direct child or snippet-based composition pattern expected by `local/no-nested-interactive`.
- Follow `wrangler.toml` and the package scripts for deployment behavior; the README deployment notes are stale.

## Checks

- `pnpm check`
- `pnpm lint`
- `pnpm format`
- `PUBLIC_API_URL=http://localhost:3000 pnpm build`
- **ALWAYS** run these after you are done making changes

## Local Patterns

- `svelte.config.js` defines the Cloudflare adapter and the `@/*` alias to `src/lib/*`.
- `src/routes/+layout.ts` sets `ssr = false`, so page load functions run in the browser and fetch backend data from `${PUBLIC_API_URL}/api/...`.
- `components.json` and `src/lib/components/ui` are the existing shadcn-svelte/Bits UI component seam; extend that before adding parallel primitives.
- The local ESLint plugin is the source of truth for the trigger/button nesting rule and its expected composition pattern.

## Useful files

- `package.json`
- `eslint.config.js`
- `svelte.config.js`
- `wrangler.toml`
- `components.json`
- `.env.example`
- `src/routes/+layout.ts`
- `eslint-plugin-local/index.js`
