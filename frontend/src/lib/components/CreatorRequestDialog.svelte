<script lang="ts">
  import { goto, invalidateAll } from '$app/navigation';
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
  let usernameInput = '';
  let loading = false;
  let error = '';
  let success = '';

  function extractUsernameFromInput(input: string): string {
    const trimmed = input.trim();

    // Check if it's a Fansly URL
    const fanslyUrlMatch = trimmed.match(/^https?:\/\/(?:www\.)?fansly\.com\/([^/?]+)/i);
    if (fanslyUrlMatch) {
      return fanslyUrlMatch[1];
    }

    // Remove @ prefix if present
    return trimmed.replace(/^@/, '');
  }

  async function handleSubmit() {
    if (!usernameInput.trim()) {
      error = 'Please enter a username or Fansly URL';
      return;
    }

    loading = true;
    error = '';
    success = '';

    try {
      const username = extractUsernameFromInput(usernameInput);

      const response = await fetch(`${PUBLIC_API_URL}/api/creators/request`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ username })
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to request creator');
      }

      success = data.message || 'Creator requested successfully';
      usernameInput = '';

      // Refresh the page data
      await invalidateAll();

      // Navigate to creators page with search
      open = false;
      success = '';
      await goto(`/creators?search=${encodeURIComponent(username)}`);
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
    <span class="hidden sm:block">Add Creator</span>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Add New Creator</DialogTitle>
      <DialogDescription>
        Enter a Fansly username to start tracking creator statistics.
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
          <Label for="username">Username or URL</Label>
          <Input
            id="username"
            placeholder="Enter username or Fansly URL (e.g. @username or https://fansly.com/username)"
            bind:value={usernameInput}
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
          {loading ? 'Adding...' : 'Add Creator'}
        </Button>
      </DialogFooter>
    </form>
  </DialogContent>
</Dialog>
