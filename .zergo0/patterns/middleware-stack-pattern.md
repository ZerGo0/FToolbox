# Middleware Stack Pattern

## When to Use
Use this pattern when configuring the Fiber application middleware stack in `main.go` to ensure consistent request processing, security, and performance optimization.

## Why It Exists
Establishes a standardized middleware order that provides proper error handling, security headers, compression, caching, and rate limiting across all API endpoints. The order is critical for correct functionality.

## Implementation
Apply middleware in this specific sequence during app initialization:

```go
app.Use(recover.New())
app.Use(cors.New(cors.Config{
    AllowOrigins: "*",
    AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
}))
app.Use(compress.New(compress.Config{
    Level: compress.LevelBestSpeed,
}))
app.Use(etag.New())
app.Use(limiter.New(limiter.Config{
    Max:        100,
    Expiration: 30 * time.Second,
    KeyGenerator: func(c *fiber.Ctx) string {
        return c.IP()
    },
}))
```

## Source References
- `backend-go/main.go:167-182` - Complete middleware stack configuration
- `backend-go/routes/routes.go:26-44` - Route-specific rate limiting extensions

## Key Conventions
- **Recovery first**: Must be first to catch panics from subsequent middleware
- **CORS second**: Handles preflight requests and headers
- **Compression third**: Compresses responses after headers are set
- **ETag fourth**: Generates cache headers before rate limiting
- **Global limiter last**: Applies to all requests after other processing
- Route-specific limiters extend the global limiter for sensitive endpoints
- Always use fiber's built-in middleware implementations for consistency