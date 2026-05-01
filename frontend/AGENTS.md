<agents_md>
<purpose>
This file provides guidance to Codex CLI when working in this part of the repository.
</purpose>

<scope>
- This file governs `frontend` and its descendants.
</scope>

<local_overview>
- `frontend` is the SvelteKit 2 + Svelte 5 client application deployed with the Cloudflare adapter.
- The app currently runs client-side only via `src/routes/+layout.ts`.
- `frontend/eslint-plugin-local` supplies the custom `local/no-nested-interactive` rule that is enabled in `eslint.config.js`.
</local_overview>

<local_rules>
- ALWAYS build backend requests from `$env/static/public` via `PUBLIC_API_URL`; do not hardcode backend hosts.
- ALWAYS reuse the existing UI primitives under `src/lib/components/ui` before introducing new component patterns.
- NEVER nest interactive controls inside trigger components such as `DialogTrigger`; use the direct child or snippet-based composition pattern expected by `local/no-nested-interactive`.
- Cloudflare Pages builds need `PUBLIC_API_URL` configured for build time; `wrangler.toml [vars]` is not enough for SvelteKit static public env resolution.
</local_rules>

<checks>
- `PUBLIC_API_URL=http://localhost:3000 pnpm check`
- `pnpm lint`
- `pnpm format`
- **ALWAYS** run these after you are done making changes
</checks>

<local_patterns>
- `svelte.config.js` defines the Cloudflare adapter and the `@/*` alias to `src/lib/*`.
- Route loaders and UI components fetch backend data from `${PUBLIC_API_URL}/api/...`.
- The local ESLint plugin is the source of truth for the trigger/button nesting rule and its expected composition pattern.
- The local package scripts are `dev`, `build`, `preview`, `prepare`, `check`, `format`, `lint`, and `deploy:production`.
</local_patterns>

<useful_files>
- `package.json`
- `eslint.config.js`
- `svelte.config.js`
- `wrangler.toml`
- `.env.example`
- `src/routes/+layout.ts`
- `eslint-plugin-local/index.js`
</useful_files>
</agents_md>
