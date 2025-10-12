# Request Validation Pattern

## When to Use
Use this pattern for validating incoming API request bodies and checking for duplicate resources before creation.

## Why It Exists
Ensures data integrity by validating request structure and preventing duplicate resource creation. Provides consistent error responses and handles existing resources gracefully.

## Implementation

### Body Parsing and Basic Validation
Parse request body and validate required fields:
```go
var req RequestStruct
if err := c.BodyParser(&req); err != nil {
    return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
}

if req.RequiredField == "" {
    return c.Status(400).JSON(fiber.Map{"error": "Field is required"})
}
```

### Duplicate Resource Check
Check for existing resources using unique identifiers:
```go
var existingModel Model
if err := h.db.Where("unique_field = ?", req.UniqueField).First(&existingModel).Error; err == nil {
    // Return existing resource like old backend
    return c.JSON(fiber.Map{
        "message": "Resource is already being tracked",
        "resource": existingModel,
    })
}
```

### Error Response Format
Use consistent 400 status codes with descriptive error messages:
```go
return c.Status(400).JSON(fiber.Map{"error": "Descriptive error message"})
```

## Source References
- `backend-go/handlers/tag_handler.go:537-543` - Tag request body parsing and validation
- `backend-go/handlers/tag_handler.go:546-552` - Tag duplicate check with existing resource response
- `backend-go/handlers/creator_handler.go:302-308` - Creator request body parsing and validation
- `backend-go/handlers/creator_handler.go:311-317` - Creator duplicate check with existing resource response

## Key Conventions
- Always use `c.BodyParser()` for JSON request body parsing
- Return 400 status for malformed request bodies
- Validate all required fields before database operations
- Check for existing resources using unique identifiers (username, tag, etc.)
- Return existing resource information instead of errors for duplicates
- Use consistent error message format: `{"error": "message"}`
- Use specific field names in error messages for clarity