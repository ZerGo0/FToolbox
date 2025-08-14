# Suggested Commands for FToolbox Development

## Frontend Development Commands

### In the `frontend` directory:
```bash
pnpm install       # Install dependencies
pnpm check         # Run svelte-check for type errors
pnpm lint          # Run Prettier and ESLint
pnpm format        # Format code with Prettier
pnpm dev           # Start development server (port 5173)
pnpm build         # Build for production
```

**IMPORTANT**: Always run `pnpm check && pnpm lint` after making frontend changes

## Backend Development Commands

### In the `backend-go` directory:
```bash
go mod download    # Install dependencies
go fmt ./...       # Format Go code
go vet ./...       # Vet code for issues
go build           # Build the application
air                # Start dev server with hot reload (port 3000)
```

**IMPORTANT**: Always run `go fmt ./...` and `go vet ./...` after making backend changes

## Task Runner Commands (from project root)
```bash
task watch-frontend  # Kill port 5173 and start frontend dev server
task watch-backend   # Kill port 3000 and start Go backend with Air
```

## Git Commands
```bash
git status         # Check current status
git diff           # View changes
git log --oneline  # View recent commits
```

## System Utilities (Linux)
```bash
ls -la             # List files with details
find . -name "*.go"  # Find files by pattern
grep -r "pattern"  # Search for text in files
ps aux | grep node # Find running processes
kill -9 $(lsof -t -i:PORT)  # Kill process on port
```