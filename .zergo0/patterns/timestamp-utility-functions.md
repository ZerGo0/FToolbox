# Timestamp Utility Functions Pattern

## When to Use
Use this pattern for converting Go time.Time values to Unix timestamps in API responses to ensure consistent timestamp handling across the frontend.

## Why It Exists
Frontend applications prefer Unix timestamps for consistent date/time handling across different time zones and browsers. This pattern provides reusable utility functions for safe timestamp conversion with proper null handling.

## Implementation

### Time Conversion Utilities
Create helper functions for timestamp conversion:
```go
func timeToUnix(t time.Time) int64 {
    return t.Unix()
}

func timeToUnixPtr(t *time.Time) *int64 {
    if t == nil {
        return nil
    }
    return ptr(t.Unix())
}

func ptr[T any](v T) *T {
    return &v
}
```

### Usage in API Responses
Convert timestamps when building response objects:
```go
return c.JSON(fiber.Map{
    "lastCheckedAt":     timeToUnixPtr(model.LastCheckedAt),
    "deletedDetectedAt": timeToUnixPtr(model.DeletedDetectedAt),
    "fanslyCreatedAt":   ptr(timeToUnix(model.FanslyCreatedAt)),
})
```

## Source References
- `backend-go/handlers/tag_handler.go:808-821` - Timestamp utility function definitions
- `backend-go/handlers/tag_handler.go:271-275` - Timestamp conversion in tag responses
- `backend-go/handlers/creator_handler.go:179-181` - Timestamp conversion in creator responses
- `backend-go/handlers/tag_handler.go:382-386` - Timestamp conversion in statistics responses

## Key Conventions
- Always use `timeToUnixPtr()` for nullable timestamp fields
- Use `timeToUnix()` for non-nullable timestamp fields
- Wrap non-nullable results with `ptr()` when response expects pointer
- Hide internal `time.Time` fields from JSON responses using `json:"-"`
- Provide Unix timestamps as int64 values for frontend consistency