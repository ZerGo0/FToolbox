<script lang="ts">
  import { goto, invalidateAll } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { PUBLIC_API_URL } from '$env/static/public';
  import { Alert, AlertDescription } from '$lib/components/ui/alert';
  import { Button, buttonVariants } from '$lib/components/ui/button';
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger
  } from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { Plus } from 'lucide-svelte';

  let open = false;
  let tagInput = '';
  let loading = false;
  let error = '';
  let success = '';

  async function handleSubmit() {
    if (!tagInput.trim()) {
      error = 'Please enter a tag name';
      return;
    }

    loading = true;
    error = '';
    success = '';

    try {
      const response = await fetch(`${PUBLIC_API_URL}/api/tags/request`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ tag: tagInput.trim().replace(/^#/, '') })
      });

      if (!response.ok) {
        const ratelimited = response.status === 429;
        if (ratelimited) {
          throw new Error('You are sending too many requests. Please try again later.');
        }

        throw new Error('Failed to request tag');
      }

      const data = await response.json();

      success = data.message || 'Tag requested successfully';
      const searchTag = tagInput.trim().replace(/^#/, '');
      tagInput = '';

      // Refresh the page data
      await invalidateAll();

      // Navigate to tags page with search
      open = false;
      success = '';
      await goto(resolve(`/tags?search=${encodeURIComponent(searchTag)}`));
    } catch (e) {
      error = e instanceof Error ? e.message : 'An error occurred';
    } finally {
      loading = false;
    }
  }
</script>

<Dialog bind:open>
  <DialogTrigger class={buttonVariants()}>
    <Plus class="h-4 w-4" />
    <span class="hidden sm:block">Add Tag</span>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Add New Tag</DialogTitle>
      <DialogDescription>
        Enter a Fansly tag name to start tracking its statistics.
      </DialogDescription>
    </DialogHeader>
    <form
      onsubmit={(e) => {
        e.preventDefault();
        handleSubmit();
      }}
    >
      <div class="space-y-4">
        <div class="space-y-2">
          <Label for="tag">Tag Name</Label>
          <Input
            id="tag"
            placeholder="Enter tag name (e.g. fyp or #fyp)"
            bind:value={tagInput}
            disabled={loading}
          />
        </div>

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
          {loading ? 'Adding...' : 'Add Tag'}
        </Button>
      </DialogFooter>
    </form>
  </DialogContent>
</Dialog>
