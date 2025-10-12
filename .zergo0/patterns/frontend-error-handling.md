# Frontend Error Handling Pattern

## When to Use
Use this pattern for all frontend API calls and async operations to ensure consistent error handling, user feedback, and graceful degradation.

## Why It Exists
Provides consistent user experience when API calls fail, prevents application crashes, and ensures users receive meaningful feedback about errors while maintaining application functionality.

## Implementation

### API Call Error Handling
Wrap API calls in try-catch with fallback data:
```typescript
try {
  const response = await fetch(`${PUBLIC_API_URL}/api/endpoint?${params}`);
  if (!response.ok) {
    throw new Error('Failed to fetch data');
  }
  const data = await response.json();
  return { data, /* other props */ };
} catch (error) {
  console.error('Error loading data:', error);
  return {
    data: [], // Fallback empty data
    pagination: { page: 1, limit: 20, totalCount: 0, totalPages: 0 },
    // other fallback props
  };
}
```

### Independent Error Handling
Handle secondary API calls independently to avoid blocking primary data:
```typescript
let statistics: StatisticsType = { /* default values */ };
try {
  const statsResponse = await fetch(`${PUBLIC_API_URL}/api/endpoint/statistics`);
  if (statsResponse.ok) {
    statistics = await statsResponse.json();
  }
} catch (statsError) {
  console.error('Error loading statistics:', statsError);
  // Continue with default values
}
```

### Form Submission Error Handling
Handle loading states, validation, and API errors:
```typescript
loading = true;
error = '';
success = '';

try {
  const response = await fetch(`${PUBLIC_API_URL}/api/endpoint`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });

  if (!response.ok) {
    const ratelimited = response.status === 429;
    if (ratelimited) {
      throw new Error('You are sending too many requests. Please try again later.');
    }
    throw new Error('Failed to process request');
  }

  const data = await response.json();
  success = data.message || 'Operation completed successfully';
  // Handle success (navigate, refresh data, etc.)
} catch (e) {
  error = e instanceof Error ? e.message : 'An error occurred';
} finally {
  loading = false;
}
```

## Source References
- `frontend/src/routes/tags/+page.ts:71-135` - Tags page API error handling
- `frontend/src/routes/creators/+page.ts:79-165` - Creators page error handling
- `frontend/src/lib/components/TagRequestDialog.svelte:35-70` - Dialog form error handling
- `frontend/src/lib/components/CreatorRequestDialog.svelte:48-85` - Creator dialog error handling
- `frontend/src/routes/banned-tags/+page.ts:40-65` - Banned tags error handling

## Key Conventions
- Always wrap API calls in try-catch blocks
- Provide meaningful fallback data for failed requests
- Log errors to console for debugging: `console.error('Context:', error)`
- Handle rate limiting (429) with specific user-friendly messages
- Use loading states to prevent duplicate submissions
- Clear previous errors before new operations
- Handle secondary API calls independently to avoid blocking primary functionality
- Always reset loading state in finally blocks
- Provide user-friendly error messages, not technical details