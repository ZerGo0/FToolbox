# API Response Structure Pattern

## When to Use
Use this pattern for all API endpoints to ensure consistent response formatting and error handling.

## Why It Exists
This pattern standardizes API responses across the application, making frontend integration predictable and reducing error handling complexity.

## Implementation Details

### Success Response Format
```go
return c.JSON(fiber.Map{
    "data": responseData,
    "pagination": fiber.Map{  // Optional for list endpoints
        "page":       page,
        "limit":      limit,
        "totalCount": total,
        "totalPages": (total + int64(limit) - 1) / int64(limit),
    },
})
```

### Error Response Format
```go
return c.Status(statusCode).JSON(fiber.Map{"error": "Error message"})
```

### Statistics Response Pattern
For statistics endpoints, return zero values when no data exists:
```go
if err == gorm.ErrRecordNotFound {
    return c.JSON(fiber.Map{
        "totalViewCount":       0,
        "totalPostCount":       0,
        "change24h":            0,
        "changePercent24h":     0,
        "calculatedAt":         nil,
    })
}
```

### Request/Response Pattern
For POST endpoints that create resources:
```go
return c.JSON(fiber.Map{
    "message": "Resource added successfully",
    "resource": createdResource,
})
```

## References
- `backend-go/handlers/tag_handler.go:403-434` - Tag statistics response
- `backend-go/handlers/creator_handler.go:257-295` - Creator statistics response
- `backend-go/handlers/tag_handler.go:609-613` - Tag request response
- `backend-go/handlers/creator_handler.go:380-384` - Creator request response