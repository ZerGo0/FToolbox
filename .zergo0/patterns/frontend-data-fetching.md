# Frontend Data Fetching Pattern

## When to Use
Use this pattern for all SvelteKit page loaders that fetch data from the Go API.

## Why It Exists
This pattern ensures consistent API integration, proper error handling, and type safety across all frontend data fetching operations.

## Implementation Details

### API URL Construction
Always use absolute URLs with `PUBLIC_API_URL`:
```typescript
import { PUBLIC_API_URL } from '$env/static/public';

const response = await fetch(`${PUBLIC_API_URL}/api/endpoint?${params}`);
```

### Parameter Handling
Extract and validate URL parameters with defaults:
```typescript
const page = url.searchParams.get('page') || '1';
const search = url.searchParams.get('search') || '';
const sortBy = url.searchParams.get('sortBy') || 'defaultField';
const sortOrder = url.searchParams.get('sortOrder') || 'desc';
```

### Date Range Defaults
Provide sensible defaults for date ranges:
```typescript
const now = new Date();
const sevenDaysAgo = new Date();
sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);

const historyStartDate = url.searchParams.get('historyStartDate') || sevenDaysAgo.toISOString();
const historyEndDate = url.searchParams.get('historyEndDate') || endOfDay.toISOString();
```

### Error Handling Pattern
Always include try-catch with fallback data:
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

### Statistics Fetching
Fetch statistics separately with independent error handling:
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

## References
- `frontend/src/routes/tags/+page.ts:51-152` - Tags page data fetching
- `frontend/src/routes/creators/+page.ts:60-165` - Creators page data fetching
- `frontend/src/routes/+layout.ts:1` - SSR disabled configuration