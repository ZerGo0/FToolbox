# Task Completion Checklist

## Frontend Tasks
After making changes to frontend code, ALWAYS:
1. Run `pnpm check` to verify no TypeScript errors
2. Run `pnpm lint` to format and lint code
3. Verify the changes work in the browser if applicable
4. Check that UI components use shadcn-svelte when available

## Backend Tasks
After making changes to backend Go code, ALWAYS:
1. Run `go fmt ./...` to format the code
2. Run `go vet ./...` to check for issues
3. Ensure database migrations are handled (GORM auto-migrate)
4. Test API endpoints if modified

## General Tasks
1. Ensure no sensitive data is exposed in code
2. Follow existing code patterns and conventions
3. Verify changes don't break existing functionality
4. Check that error handling is appropriate
5. Ensure proper logging is in place

## Before Committing (only when explicitly asked)
1. All linting and formatting commands pass
2. No type errors or compilation issues
3. Changes are tested and working
4. No secrets or sensitive data included