# Go Handler Pattern

## When to Use
Use this pattern when creating API endpoints that handle HTTP requests with consistent error handling, pagination, and response formatting.

## Why It Exists
This pattern ensures consistent API behavior across all endpoints, standardizes error responses, and provides reusable pagination and sorting logic.

## Implementation
Handler structs contain dependencies (database, clients) and implement methods for each HTTP endpoint. Key characteristics:

- Dependency injection via constructor functions (`NewCreatorHandler`, `NewTagHandler`)
- Consistent error handling with Zap logging and JSON error responses
- Standardized pagination with `page`, `limit`, and total count calculations
- Query parameter mapping for sorting (`sortBy`, `sortOrder` to database columns)
- Optional history inclusion with date range filtering
- Batch loading for related data to avoid N+1 queries

## References
- `backend-go/handlers/creator_handler.go:15-25` - Handler struct with dependency injection
- `backend-go/handlers/creator_handler.go:57-72` - Pagination parameter parsing and validation
- `backend-go/handlers/creator_handler.go:95-104` - Column mapping for sorting
- `backend-go/handlers/tag_handler.go:85-128` - Complex query parameter handling with tag parsing
- `backend-go/handlers/tag_handler.go:119-122` - Pagination limits and validation
- `backend-go/handlers/worker_handler.go:19-50` - Simple handler pattern for status endpoints

## Key Conventions
- Always use `zap.L().Error()` for error logging with context
- Return JSON responses with `fiber.Map{"error": "message"}` format
- Clamp pagination limits (1-100) to prevent abuse
- Use `timeToUnixPtr()` helper for consistent timestamp formatting
- Apply rate limiting to sensitive endpoints like creation requests