# API Handler Pagination Pattern

## When to Use
Use this pattern for any list endpoint that supports pagination, sorting, searching, and optional history data.

## Why It Exists
This pattern provides consistent pagination and query handling across all list endpoints, ensuring predictable API behavior and reducing code duplication.

## Implementation Details

### Query Parameter Parsing
```go
page, _ := strconv.Atoi(c.Query("page", "1"))
limit, _ := strconv.Atoi(c.Query("limit", "20"))
search := c.Query("search")
sortBy := c.Query("sortBy", "viewCount")
sortOrder := c.Query("sortOrder", "desc")
includeHistory := c.Query("includeHistory") == "true"
```

### Parameter Validation
```go
if page < 1 {
    page = 1
}
if limit < 1 || limit > 100 {
    limit = 20
}
```

### Column Mapping
Map frontend field names to database columns for consistent sorting:
```go
columnMap := map[string]string{
    "viewCount": "view_count",
    "postCount": "post_count",
    "updatedAt": "updated_at",
    "tag":       "tag",
    "rank":      "rank",
}
```

### Pagination Response Format
```go
return c.JSON(fiber.Map{
    "items": items,
    "pagination": fiber.Map{
        "page":       page,
        "limit":      limit,
        "totalCount": total,
        "totalPages": (total + int64(limit) - 1) / int64(limit),
    },
})
```

## References
- `backend-go/handlers/tag_handler.go:85-131` - Tag pagination implementation
- `backend-go/handlers/creator_handler.go:57-89` - Creator pagination implementation
- `backend-go/handlers/tag_handler.go:436-489` - Banned tags pagination with special sorting