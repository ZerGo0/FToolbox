# Code Style and Conventions

## Frontend (SvelteKit/TypeScript)

### Formatting
- **Indentation**: 2 spaces (no tabs)
- **Quotes**: Single quotes
- **Trailing Comma**: None
- **Print Width**: 100 characters
- **Prettier**: Configured with prettier-plugin-svelte and prettier-plugin-tailwindcss

### Conventions
- **Components**: Use shadcn-svelte components from `frontend/src/lib/components/ui/` when possible
- **File Structure**: 
  - Routes in `/src/routes/`
  - Components in `/src/lib/components/`
  - UI components in `/src/lib/components/ui/`
  - Utilities in `/src/lib/utils.ts`
  - Types in `/src/lib/fanslyTypes.ts`
- **Svelte 5**: Use runes and `.svelte.ts` files
- **ESLint**: Custom rule to prevent nested interactive elements
- **TypeScript**: Strict typing enabled

## Backend (Go)

### Formatting
- **Standard Go formatting**: Use `go fmt`
- **Vetting**: Use `go vet` for code issues

### Project Structure
```
backend-go/
├── config/         # Environment configuration
├── database/       # DB connection and migrations
├── models/         # GORM models
├── handlers/       # HTTP request handlers
├── routes/         # Route definitions
├── workers/        # Background job implementations
├── fansly/         # External API client
├── ratelimit/      # Rate limiting utilities
└── utils/          # Helper functions
```

### Conventions
- **ORM**: GORM with auto-migration on startup
- **Logging**: Use Zap logger
- **Error Handling**: Return structured errors in API responses
- **Database**: MariaDB with GORM models
- **API**: RESTful endpoints under `/api` prefix

## General
- **No emojis** in code unless explicitly requested
- **Comments**: Minimal, only when necessary for complex logic
- **Security**: Never expose or log secrets/keys
- **Git**: Never commit secrets or sensitive data