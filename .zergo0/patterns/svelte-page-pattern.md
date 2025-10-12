# Svelte Page Pattern

## When to Use
Use this pattern for all page components that display tabular data with search, sorting, pagination, and optional history visualization.

## Why It Exists
This pattern provides consistent user experience across data-heavy pages with standardized interactions, responsive design, and efficient data loading.

## Implementation
Pages use Svelte 5 syntax with shared state management and URL synchronization. Key characteristics:

- Data loading via `+page.ts` with URL parameter parsing
- State management with `$state` for local UI state
- URL synchronization using `SvelteURLSearchParams`
- Consistent table layout with expandable rows for history
- Date range picker with presets and manual application
- Loading states with toast notifications for long operations
- Responsive pagination with mobile-friendly controls
- Search functionality with hashtag parsing for tags

## References
- `frontend/src/routes/creators/+page.svelte:47-85` - State initialization and date range handling
- `frontend/src/routes/creators/+page.svelte:126-147` - Search and sort parameter handling
- `frontend/src/routes/creators/+page.svelte:393-598` - Table implementation with expandable rows
- `frontend/src/routes/tags/+page.svelte:127-145` - Hashtag parsing for tag search
- `frontend/src/routes/tags/+page.svelte:577-708` - Pagination with mobile-responsive design
- `frontend/src/routes/+layout.ts:1` - SSR disabled for client-side routing

## Key Conventions
- Use `$props()` for data passed from `+page.ts`
- Implement loading toasts for operations over 1 second
- Use `formatNumber()` and `formatDate()` utilities for consistent display
- Apply responsive design with mobile-specific UI adjustments
- Include data notices for incomplete/growing datasets
- Use shadcn-svelte components for consistent UI elements
- Implement keyboard navigation (Enter to search, Escape to cancel)