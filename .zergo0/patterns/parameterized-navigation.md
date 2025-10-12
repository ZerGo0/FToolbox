# Parameterized Navigation Pattern

## When to Use
Use this pattern for client-side navigation that updates URL parameters to reflect current filter, sort, and pagination state.

## Why It Exists
Maintains application state in URL parameters for bookmarkable, shareable links and proper browser history management. Ensures UI state is preserved across page refreshes and navigation.

## Implementation

### Parameter Construction
Build URL parameter strings from current state:
```svelte
function updateParams() {
  const params = new URLSearchParams();
  
  if (data.search) params.set('search', data.search);
  if (data.sortBy !== defaultSortBy) params.set('sortBy', data.sortBy);
  if (data.sortOrder !== defaultSortOrder) params.set('sortOrder', data.sortOrder);
  if (data.page !== 1) params.set('page', data.page.toString());
  
  return params.toString();
}
```

### Navigation Trigger
Navigate with updated parameters:
```svelte
async function handleSortChange(newSortBy: string) {
  data.sortBy = newSortBy;
  await goto(`?${updateParams()}`);
}

async function handleSearch(newSearch: string) {
  data.search = newSearch;
  data.page = 1; // Reset to first page
  await goto(`?${updateParams()}`);
}
```

### Parameter Extraction
Extract parameters from URL on page load:
```svelte
const page = url.searchParams.get('page') || '1';
const search = url.searchParams.get('search') || '';
const sortBy = url.searchParams.get('sortBy') || 'defaultField';
const sortOrder = url.searchParams.get('sortOrder') || 'desc';
```

## Source References
- `frontend/src/routes/tags/+page.svelte:144,159,165,223` - Tag page parameterized navigation
- `frontend/src/routes/creators/+page.svelte:131,146,152,210` - Creator page parameterized navigation
- `frontend/src/routes/banned-tags/+page.svelte:81,96,102` - Banned tags parameterized navigation
- `frontend/src/lib/components/TagRequestDialog.svelte:65` - Dialog navigation with search parameter
- `frontend/src/lib/components/CreatorRequestDialog.svelte:79` - Dialog navigation with username parameter

## Key Conventions
- Use `URLSearchParams` for proper parameter encoding
- Include only non-default parameters in URLs to keep them clean
- Reset pagination to page 1 when filters change
- Use `goto()` from `$app/navigation` for client-side navigation
- Preserve parameter state across user interactions
- Handle search terms with proper encoding using `encodeURIComponent`
- Maintain browser history for proper back/forward navigation
- Use consistent parameter names across pages (search, sortBy, sortOrder, page)