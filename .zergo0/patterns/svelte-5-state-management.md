# Svelte 5 State Management Pattern

## When to Use
Use this pattern for managing component state in Svelte 5 applications using the new `$state` rune for reactive state management.

## Why It Exists
Svelte 5 introduces `$state` runes as the modern way to manage reactive component state, replacing the traditional `let` declarations with automatic reactivity. This pattern ensures consistent state management across all components.

## Implementation

### Basic State Declaration
Use `$state` for reactive component state:
```svelte
<script lang="ts">
  let expandedId = $state<string | null>(null);
  let searchValue = $state(data.search || '');
  let loading = $state(false);
  let error = $state<string | null>(null);
  
  // Complex objects
  let dateRangeValue = $state<DateRange>({
    start: today(getLocalTimeZone()).subtract({ days: 7 }),
    end: today(getLocalTimeZone())
  });
</script>
```

### State Updates
State updates automatically trigger reactivity:
```svelte
// Direct assignment
loading = true;
error = null;

// Object property updates
dateRangeValue = { start: newStart, end: newEnd };
```

### Effect Hooks
Use `$effect` for side effects based on state changes:
```svelte
$effect(() => {
  if (data.historyStartDate && data.historyEndDate) {
    dateRangeValue = {
      start: parseDate(data.historyStartDate),
      end: parseDate(data.historyEndDate)
    };
  }
});
```

## Source References
- `frontend/src/routes/tags/+page.svelte:51-62` - Page-level state management
- `frontend/src/routes/creators/+page.svelte:50-58` - Creator page state
- `frontend/src/lib/components/PostTextDialog.svelte:18-32` - Dialog component state
- `frontend/src/lib/components/AppSidebar.svelte:22-23` - Sidebar state management

## Key Conventions
- Always use `$state` for reactive component state instead of `let`
- Use TypeScript types for state variables: `$state<Type>()`
- Initialize state with sensible defaults or props data
- Use `$effect` for side effects that depend on state changes
- State updates are automatically reactive, no need for `$:` in Svelte 5
- Keep state close to where it's used to minimize prop drilling