# Chart Component Pattern

## When to Use
Use this pattern for creating reusable chart components that display time-series data with consistent styling, responsive design, and proper data transformation.

## Why It Exists
This pattern provides consistent chart visualization across the application with standardized data handling, responsive design, and proper cleanup to prevent memory leaks.

## Implementation
Chart components use Chart.js with Svelte 5 reactivity and proper lifecycle management. Key characteristics:

- Props-based data input with optional history arrays
- Automatic data transformation for timestamp handling
- Chart instance management with proper cleanup
- Responsive design with consistent color schemes
- Multi-dataset support with different Y-axes
- Empty state handling with graceful degradation
- Effect-based updates to prevent unnecessary re-renders

## References
- `frontend/src/lib/components/TagHistory.svelte:19-27` - Props interface and chart state management
- `frontend/src/lib/components/TagHistory.svelte:28-41` - Data transformation with timestamp conversion
- `frontend/src/lib/components/TagHistory.svelte:43-48` - Effect-based chart updates
- `frontend/src/lib/components/TagHistory.svelte:60-91` - Dataset preparation with empty data handling
- `frontend/src/lib/components/CreatorHistory.svelte:19-27` - Similar structure for creator history
- `frontend/src/lib/components/CreatorHistory.svelte:54-96` - Multi-dataset configuration with consistent styling

## Key Conventions
- Use $props() for data input and $state() for chart instance management
- Transform timestamps to Date objects consistently across components
- Implement proper chart cleanup in effects to prevent memory leaks
- Use consistent color schemes and styling across all charts
- Handle empty data gracefully with fallback empty datasets
- Use setTimeout to defer chart updates until DOM is ready
- Support multiple datasets with different Y-axes for complex data
- Apply responsive design with proper aspect ratios