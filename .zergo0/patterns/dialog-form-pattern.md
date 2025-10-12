# Dialog Form Pattern

## When to Use
Use this pattern for modal dialogs that collect user input for creating new entities (creators, tags) with proper validation and error handling.

## Why It Exists
This pattern provides consistent user experience for data entry with proper validation, loading states, and seamless integration with the main data views.

## Implementation
Dialogs use shadcn-svelte Dialog components with form handling. Key characteristics:

- Controlled open state with proper binding
- Input validation with user-friendly error messages
- Loading states during API calls
- Success feedback with automatic navigation
- URL parsing for entity extraction (e.g., Fansly URLs)
- Rate limiting awareness with appropriate error messages
- Form reset and dialog closure on success

## References
- `frontend/src/lib/components/CreatorRequestDialog.svelte:25-36` - Username extraction from URLs
- `frontend/src/lib/components/CreatorRequestDialog.svelte:38-85` - Form submission with error handling
- `frontend/src/lib/components/TagRequestDialog.svelte:25-71` - Simplified tag form handling
- `frontend/src/lib/components/CreatorRequestDialog.svelte:88-138` - Dialog structure with validation
- `frontend/src/lib/components/TagRequestDialog.svelte:74-124` - Consistent dialog layout

## Key Conventions
- Use `bind:open` for dialog state control
- Implement client-side validation before API calls
- Handle HTTP 429 status with specific rate limit messages
- Use `invalidateAll()` to refresh data after creation
- Navigate to entity page with search parameters after success
- Reset form state and close dialog on successful submission
- Provide clear feedback for both errors and success states