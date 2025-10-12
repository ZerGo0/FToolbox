# HTTP Client with Rate Limit Pattern

## When to Use
Use this pattern for external API clients that need to respect rate limits, handle retries, and provide consistent error handling for HTTP requests.

## Why It Exists
This pattern provides robust external API integration with automatic rate limiting, retry logic, proper error handling, and context-based cancellation to prevent API abuse and ensure reliability.

## Implementation
HTTP clients wrap standard HTTP functionality with rate limiting and retry logic. Key characteristics:

- Global rate limiting with configurable windows and request limits
- Automatic retry with exponential backoff for transient failures
- Context-based cancellation for timeout handling
- Proper header management including User-Agent and Authorization
- Structured error handling with specific error types
- Request spreading to avoid burst patterns
- Comprehensive logging for debugging and monitoring

## References
- `backend-go/fansly/client.go:27-52` - Client struct with rate limiter and HTTP client
- `backend-go/fansly/client.go:61-108` - Request execution with rate limiting and headers
- `backend-go/fansly/client.go:109-157` - Retry logic with exponential backoff
- `backend-go/ratelimit/global.go:11-28` - Global rate limiter implementation
- `backend-go/ratelimit/global.go:30-58` - Wait method with request cleanup and blocking
- `backend-go/ratelimit/global.go:84-107` - Request spreading for even distribution

## Key Conventions
- Always use context.Context for cancellation and timeout handling
- Implement global rate limiting before making requests
- Use exponential backoff for retries with maximum attempts
- Set proper User-Agent headers for API identification
- Handle HTTP 429 status codes with specific retry logic
- Log rate limit events for monitoring and debugging
- Spread requests evenly to avoid burst patterns
- Return structured error types for different failure scenarios