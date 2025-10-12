# HTTP Status Code Constants Pattern

## When to Use
Use this pattern for HTTP status codes in API handlers to ensure consistent error responses and maintainable code.

## Why It Exists
Using Fiber's built-in status code constants provides type safety, improves code readability, and prevents magic numbers in error handling. This ensures consistent HTTP status code usage across all API endpoints.

## Implementation

### Error Response Pattern
Always use Fiber status constants with consistent JSON error format:
```go
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error message"})
return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Rate limit exceeded"})
```

### Request Validation
Use appropriate status codes for validation failures:
```go
if err := c.BodyParser(&req); err != nil {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
}

if req.RequiredField == "" {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Field is required"})
}
```

### Resource Not Found
Consistent 404 handling for missing resources:
```go
if err == gorm.ErrRecordNotFound {
    return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
}
```

## Source References
- `backend-go/handlers/tag_handler.go:420` - Internal server error response
- `backend-go/handlers/tag_handler.go:621` - Bad request for missing query params
- `backend-go/handlers/tag_handler.go:667` - Not found for missing data
- `backend-go/handlers/creator_handler.go:277` - Internal server error pattern
- `backend-go/main.go:156` - Global error handler with status constants

## Key Conventions
- Always use `fiber.Status*` constants instead of numeric codes
- Maintain consistent JSON error format: `{"error": "message"}`
- Use 400 for client validation errors
- Use 404 for missing resources
- Use 429 for rate limit exceeded
- Use 500 for unexpected server errors
- Log detailed errors server-side, return generic messages to clients