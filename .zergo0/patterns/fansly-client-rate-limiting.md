# Fansly Client Rate Limiting and Retry Pattern

## When to Use
Use this pattern for any external API client that needs rate limiting, retry logic, and proper error handling.

## Why It Exists
This pattern prevents API abuse, handles transient failures gracefully, and provides consistent behavior across all external API interactions.

## Implementation Details

### Client Structure
```go
type Client struct {
    httpClient    *http.Client
    globalLimiter *ratelimit.GlobalRateLimiter
    authToken     string
    logger        *zap.Logger
}
```

### Rate Limiting Integration
Apply global rate limiting before each request:
```go
if c.globalLimiter != nil {
    if err := c.globalLimiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limiter error: %w", err)
    }
}
```

### Retry Logic
Implement exponential backoff for retryable status codes:
```go
retryableStatuses := map[int]bool{
    http.StatusTooManyRequests:     true,
    http.StatusInternalServerError: true,
    http.StatusBadGateway:          true,
    http.StatusServiceUnavailable:  true,
    http.StatusGatewayTimeout:      true,
}

maxRetries := 3
backoff := time.Second

for attempt := 0; attempt <= maxRetries; attempt++ {
    resp, err = c.httpClient.Do(req.Clone(ctx))
    if retryableStatuses[resp.StatusCode] && attempt < maxRetries {
        // Special handling for rate limits
        if resp.StatusCode == http.StatusTooManyRequests {
            backoff = 30 * time.Second
        }
        
        select {
        case <-time.After(backoff):
            backoff *= 2
            continue
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    break
}
```

### Request Context Support
All methods accept context for cancellation:
```go
func (c *Client) GetTagWithContext(ctx context.Context, tagName string) (*TagResponseData, error) {
    // Use ctx in all HTTP requests
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
}
```

### Error Handling Patterns
Define specific error types for common failures:
```go
var ErrTagNotFound = errors.New("fansly: tag not found")

if !response.Success || response.Response == nil {
    return nil, ErrTagNotFound
}
```

### Configuration Pattern
Configure rate limits from environment:
```go
func (c *Client) SetGlobalRateLimit(maxRequests int, windowSeconds int) {
    c.globalLimiter = ratelimit.NewGlobalRateLimiter(maxRequests, windowSeconds, c.logger)
}
```

## References
- `backend-go/fansly/client.go:27-52` - Client structure and initialization
- `backend-go/fansly/client.go:61-157` - Request execution with rate limiting and retries
- `backend-go/fansly/client.go:182-204` - Context-aware API method
- `backend-go/config/config.go:43-44` - Rate limit configuration