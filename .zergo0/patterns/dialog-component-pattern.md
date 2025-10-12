# Dialog Component Pattern

## When to Use
Use this pattern for any modal dialog that handles form submissions with loading states, error handling, and navigation.

## Why It Exists
This pattern provides consistent user experience for all modal interactions, including proper loading states, error handling, and post-submission navigation.

## Implementation Details

### Component Structure
```svelte
<script lang="ts">
  import { goto, invalidateAll } from '$app/navigation';
  import { PUBLIC_API_URL } from '$env/static/public';
  // UI imports...
  
  let open = false;
  let input = '';
  let loading = false;
  let error = '';
  let success = '';
</script>
```

### Form Submission Pattern
```typescript
async function handleSubmit() {
  if (!input.trim()) {
    error = 'Please enter a value';
    return;
  }

  loading = true;
  error = '';
  success = '';

  try {
    const response = await fetch(`${PUBLIC_API_URL}/api/endpoint`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ field: processedInput })
    });

    if (!response.ok) {
      const ratelimited = response.status === 429;
      if (ratelimited) {
        throw new Error('You are sending too many requests. Please try again later.');
      }
      throw new Error('Failed to request resource');
    }

    const data = await response.json();
    success = data.message || 'Resource requested successfully';
    input = '';

    // Refresh data and navigate
    await invalidateAll();
    open = false;
    success = '';
    await goto(`/target?search=${encodeURIComponent(searchTerm)}`);
  } catch (e) {
    error = e instanceof Error ? e.message : 'An error occurred';
  } finally {
    loading = false;
  }
}
```

### Input Processing Pattern
Extract and clean user input:
```typescript
function extractFromInput(input: string): string {
  const trimmed = input.trim();
  
  // Handle URL patterns
  const urlMatch = trimmed.match(/^https?:\/\/(?:www\.)?domain\.com\/([^/?]+)/i);
  if (urlMatch) {
    return urlMatch[1];
  }
  
  // Remove prefixes
  return trimmed.replace(/^[@#]/, '');
}
```

### Template Structure
```svelte
<Dialog bind:open>
  <DialogTrigger class={buttonVariants()}>
    <Icon class="h-4 w-4" />
    <span class="hidden sm:block">Action Text</span>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Dialog Title</DialogTitle>
      <DialogDescription>
        Description of the action.
      </DialogDescription>
    </DialogHeader>
    <form onsubmit|preventDefault={handleSubmit}>
      <div class="space-y-4">
        <!-- Form fields -->
        
        {#if error}
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        {/if}
        
        {#if success}
          <Alert>
            <AlertDescription>{success}</AlertDescription>
          </Alert>
        {/if}
      </div>
      
      <DialogFooter class="mt-6">
        <Button type="button" variant="outline" onclick={() => (open = false)}>Cancel</Button>
        <Button type="submit" disabled={loading}>
          {loading ? 'Processing...' : 'Submit'}
        </Button>
      </DialogFooter>
    </form>
  </DialogContent>
</Dialog>
```

## References
- `frontend/src/lib/components/TagRequestDialog.svelte:25-71` - Tag dialog submission pattern
- `frontend/src/lib/components/CreatorRequestDialog.svelte:38-85` - Creator dialog with URL parsing
- `frontend/src/lib/components/TagRequestDialog.svelte:74-124` - Dialog template structure