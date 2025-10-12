# Navigation Loading Toast Pattern

## When to Use
Use this pattern for page components that need to show loading feedback during navigation operations that may take time to complete.

## Why It Exists
Provides user feedback during navigation operations (sorting, filtering, pagination) by showing a loading toast after a 1-second delay. This prevents brief flashes of loading states while still informing users about longer operations.

## Implementation

### Navigation State Tracking
Track navigation state and manage toast lifecycle:
```svelte
<script lang="ts">
  import { toast } from 'svelte-sonner';
  import { navigating } from '$app/stores';
  
  let loadingTimeout: number | undefined;
  let loadingToastId: string | number | undefined;
</script>
```

### Loading Toast Effect
Show loading toast after 1-second delay with proper cleanup:
```svelte
$effect(() => {
  if ($navigating) {
    // Start a timer to show loading toast after 1 second
    loadingTimeout = setTimeout(() => {
      loadingToastId = toast.loading('Loading content...');
    }, 1000);
  } else {
    // Clear timeout and dismiss loading toast when navigation completes
    if (loadingTimeout) {
      clearTimeout(loadingTimeout);
      loadingTimeout = undefined;
    }
    if (loadingToastId !== undefined) {
      toast.dismiss(loadingToastId);
      loadingToastId = undefined;
    }
  }

  return () => {
    // Cleanup timeout on effect destroy
    if (loadingTimeout) {
      clearTimeout(loadingTimeout);
    }
    if (loadingToastId !== undefined) {
      toast.dismiss(loadingToastId);
    }
  };
});
```

## Source References
- `frontend/src/routes/tags/+page.svelte:88-115` - Tags page navigation loading toast
- `frontend/src/routes/creators/+page.svelte:87-115` - Creators page navigation loading toast  
- `frontend/src/routes/banned-tags/+page.svelte:46-74` - Banned tags navigation loading toast

## Key Conventions
- Use 1-second delay before showing loading toast to avoid brief flashes
- Always clean up timeouts and toast dismissals in effect cleanup
- Track both timeout ID and toast ID for proper cleanup
- Use descriptive loading messages specific to the page content
- Import `navigating` store from `$app/stores` for navigation state
- Use `svelte-sonner` for consistent toast styling across the application