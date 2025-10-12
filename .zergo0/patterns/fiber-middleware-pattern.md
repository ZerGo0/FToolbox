# Fiber Middleware Pattern

## When to Use
Use this pattern for configuring HTTP middleware stack in Fiber applications to handle cross-cutting concerns like CORS, compression, rate limiting, and error handling.

## Why It Exists
This pattern ensures consistent request processing, security, and performance optimization across all API endpoints.

## Implementation
Middleware is configured in main.go with proper ordering and error handling. Key characteristics:

- Global error handler with consistent JSON responses
- CORS configuration for cross-origin requests
- Compression for response size optimization
- ETag support for conditional requests
- Rate limiting with IP-based key generation
- Recovery middleware for panic handling
- Proper middleware ordering (recovery first, then business logic)

## References
- `backend-go/main.go:154-164` - Global error handler configuration
- `backend-go/main.go:166-172` - CORS middleware setup
- `backend-go/main.go:172-182` - Compression, ETag, and rate limiting
- `backend-go/routes/routes.go:26-33` - Route-specific rate limiting
- `backend-go/routes/routes.go:38-45` - Rate limiting for creator requests
- `backend-go/fansly/client.go:62-157` - Retry logic with rate limit handling

## Key Conventions
- Apply recovery middleware first to catch panics
- Use X-Forwarded-For header for rate limiting behind proxies
- Configure CORS for development (allow all origins)
- Apply compression before other middleware for efficiency
- Use consistent error response format across all endpoints
- Implement route-specific rate limiting for sensitive operations
- Handle HTTP 429 status codes with appropriate retry logic