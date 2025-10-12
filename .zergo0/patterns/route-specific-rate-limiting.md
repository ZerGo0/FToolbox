# Route-Specific Rate Limiting Pattern

## When to Use
Use this pattern for applying stricter rate limits to sensitive API endpoints (POST routes, data modification operations) that require additional protection beyond the global rate limiter.

## Why It Exists
Provides targeted protection for endpoints that could be abused or resource-intensive, while maintaining more lenient limits for general read operations. Prevents brute force attacks and resource exhaustion on critical routes.

## Implementation
Create per-endpoint limiters with X-Forwarded-For support:

```go
// Create route-specific limiter
postLimiter := limiter.New(limiter.Config{
    Max:        5,
    Expiration: 60 * time.Second,
    KeyGenerator: func(c *fiber.Ctx) string {
        // Support proxy environments
        if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
            return forwarded
        }
        return c.IP()
    },
    LimitReached: func(c *fiber.Ctx) error {
        return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
            "error": "Rate limit exceeded",
        })
    },
})

// Apply to specific route
app.Post("/api/creators", postLimiter, handler.CreateCreator)
```

## Source References
- `backend-go/routes/routes.go:26-44` - Route-specific rate limiting implementation
- `backend-go/routes/routes.go:31` - POST endpoint with stricter limits
- `backend-go/routes/routes.go:38` - Error response for rate limit exceeded

## Key Conventions
- Use stricter limits for POST/PUT/DELETE operations (typically 5-10 requests per minute)
- Always check `X-Forwarded-For` header for proxy environments
- Return consistent `429 Too Many Requests` with JSON error format
- Apply limiter middleware before the route handler
- Use descriptive limiter variable names (e.g., `postLimiter`, `uploadLimiter`)
- Keep expiration times reasonable (60 seconds for most cases)