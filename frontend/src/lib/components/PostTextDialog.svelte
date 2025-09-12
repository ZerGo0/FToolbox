<script lang="ts">
  import { PUBLIC_API_URL } from '$env/static/public';
  import { Badge } from '$lib/components/ui/badge';
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
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from '$lib/components/ui/textarea';
  import { Clipboard, Hash, Loader2, TrendingDown, TrendingUp } from 'lucide-svelte';

  let open = $state(false);
  let postText = $state('');
  let extractedTags = $state<
    Array<{
      tag: string;
      viewCount?: number | null;
      rank?: number | null;
      history?: Array<{ viewCount: number }>;
      exists?: boolean;
      loading?: boolean;
      error?: string | undefined;
    }>
  >([]);
  let loading = $state(false);
  let error = $state<string | null>(null);

  function extractTagsFromText(text: string): string[] {
    // Match hashtags (words starting with # followed by alphanumeric characters)
    const tagRegex = /#(\w+)/g;
    const matches = text.matchAll(tagRegex);
    const tags = Array.from(matches, (m) => m[1].toLowerCase());
    // Remove duplicates
    return [...new Set(tags)];
  }

  async function fetchTagStats(tags: string[]) {
    if (tags.length === 0) return [];

    // Initialize all tags with loading state
    const results: Array<{
      tag: string;
      viewCount: number | null;
      rank: number | null;
      exists: boolean;
      loading: boolean;
      error: string | undefined;
    }> = tags.map((tag) => ({
      tag,
      viewCount: null,
      rank: null,
      exists: false,
      loading: true,
      error: undefined
    }));

    // Fetch stats for each tag using the request endpoint
    const promises = tags.map(async (tag, index) => {
      try {
        const response = await fetch(`${PUBLIC_API_URL}/api/tags/request`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ tag })
        });

        if (!response.ok) {
          const ratelimited = response.status === 429;
          if (ratelimited) {
            throw new Error('You are sending too many requests. Please try again later.');
          }

          throw new Error('Failed to request tag');
        }

        const data = await response.json();

        if (response.ok && data.tag) {
          results[index] = {
            tag: data.tag.tag,
            viewCount: data.tag.viewCount,
            rank: data.tag.rank,
            exists: true,
            loading: false,
            error: undefined
          };
        } else if (response.status === 404) {
          results[index] = {
            tag,
            viewCount: null,
            rank: null,
            exists: false,
            loading: false,
            error: 'Tag not found on Fansly'
          };
        } else {
          results[index] = {
            tag,
            viewCount: null,
            rank: null,
            exists: false,
            loading: false,
            error: data.error || 'Failed to fetch tag'
          };
        }
      } catch {
        results[index] = {
          tag,
          viewCount: null,
          rank: null,
          exists: false,
          loading: false,
          error: 'Network error'
        };
      }
    });

    await Promise.all(promises);
    return results;
  }

  async function analyzePostText() {
    error = null;

    if (!postText.trim()) {
      error = 'Please enter some text to analyze';
      return;
    }

    loading = true;

    try {
      // Extract tags from text
      const tags = extractTagsFromText(postText);

      if (tags.length === 0) {
        error = 'No hashtags found in the text. Hashtags should start with #';
        extractedTags = [];
        return;
      }

      // Fetch stats for extracted tags
      const tagStats = await fetchTagStats(tags);
      // Sort by view count (highest first)
      extractedTags = tagStats.sort((a, b) => {
        // Put tags with view counts before those without
        if (a.viewCount === null && b.viewCount !== null) return 1;
        if (a.viewCount !== null && b.viewCount === null) return -1;
        if (a.viewCount === null && b.viewCount === null) return 0;
        // Sort by view count descending
        return (b.viewCount ?? 0) - (a.viewCount ?? 0);
      });
    } catch (err) {
      error = 'Failed to analyze text';
      console.error(err);
    } finally {
      loading = false;
    }
  }

  function formatNumber(num: number | null | undefined): string {
    if (num == null) return 'N/A';
    return new Intl.NumberFormat().format(num);
  }

  async function requestTag(tag: string, index: number) {
    // Update loading state for this specific tag
    extractedTags[index] = { ...extractedTags[index], loading: true, error: undefined };

    try {
      const response = await fetch('http://localhost:3000/api/tags/request', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ tag })
      });

      if (!response.ok) {
        const ratelimited = response.status === 429;
        if (ratelimited) {
          throw new Error('You are sending too many requests. Please try again later.');
        }
        throw new Error('Failed to request tag');
      }

      const data = await response.json();

      if (response.ok && data.tag) {
        extractedTags[index] = {
          tag: data.tag.tag,
          viewCount: data.tag.viewCount,
          rank: data.tag.rank,
          exists: true,
          loading: false,
          error: undefined
        };
      } else {
        extractedTags[index] = {
          ...extractedTags[index],
          loading: false,
          error: data.error || 'Failed to add tag'
        };
      }
    } catch {
      extractedTags[index] = {
        ...extractedTags[index],
        loading: false,
        error: 'Network error'
      };
    }
  }

  function getTagChange(tag: { history?: Array<{ viewCount: number }> }) {
    if (!tag.history || tag.history.length === 0) {
      return { change: 0, percentage: 0 };
    }

    // History is in descending order (newest first)
    const newestViewCount = tag.history[0].viewCount;
    const oldestViewCount = tag.history[tag.history.length - 1].viewCount;
    const totalChange = newestViewCount - oldestViewCount;
    const percentage = oldestViewCount > 0 ? (totalChange / oldestViewCount) * 100 : 0;

    return { change: totalChange, percentage };
  }

  function reset() {
    postText = '';
    extractedTags = [];
    error = null;
  }

  $effect(() => {
    if (!open) {
      reset();
    }
  });
</script>

<Dialog bind:open>
  <DialogTrigger class={buttonVariants({ variant: 'outline' })}>
    <Clipboard class="h-4 w-4" />
    <span class="hidden sm:block">Analyze Post Text</span>
  </DialogTrigger>
  <DialogContent class="max-w-2xl">
    <DialogHeader>
      <DialogTitle>Analyze Post Text</DialogTitle>
      <DialogDescription>
        Paste your post text to extract hashtags and view their statistics
      </DialogDescription>
    </DialogHeader>

    <div class="grid gap-4 py-4">
      <div class="grid gap-2">
        <Label for="post-text">Post Text</Label>
        <Textarea
          id="post-text"
          bind:value={postText}
          placeholder="Enter your post text with #hashtags..."
          class="min-h-[120px]"
        />
      </div>

      {#if error}
        <div class="bg-destructive/10 text-destructive rounded-md p-3 text-sm">
          {error}
        </div>
      {/if}

      {#if extractedTags.length > 0}
        <div class="space-y-3">
          <h4 class="text-sm font-medium">Extracted Tags ({extractedTags.length})</h4>
          <div class="max-h-[300px] space-y-2 overflow-y-auto rounded-md border p-3">
            {#each extractedTags as tag, index (tag.tag)}
              {@const changeData = getTagChange(tag)}
              <div class="bg-muted/50 flex items-center justify-between rounded-lg p-3">
                <div class="flex items-center gap-2">
                  <Hash class="text-muted-foreground h-4 w-4" />
                  <Badge variant="secondary">#{tag.tag}</Badge>
                </div>
                <div class="flex items-center gap-4 text-sm">
                  {#if tag.loading}
                    <Loader2 class="h-4 w-4 animate-spin" />
                  {:else if tag.error}
                    <span class="text-red-500">{tag.error}</span>
                    {#if tag.error === 'Tag not found on Fansly'}
                      <span class="text-muted-foreground text-xs">(Not available)</span>
                    {/if}
                  {:else if tag.exists}
                    {#if tag.rank !== null}
                      <span class="text-muted-foreground">Rank #{tag.rank}</span>
                    {/if}
                    {#if tag.viewCount !== null}
                      <span class="font-medium">{formatNumber(tag.viewCount)} views</span>
                    {/if}
                    {#if changeData.change !== 0}
                      <div class="flex items-center gap-1">
                        {#if changeData.change > 0}
                          <TrendingUp class="h-3 w-3 text-green-500" />
                          <span class="text-xs text-green-500">
                            +{changeData.percentage.toFixed(1)}%
                          </span>
                        {:else}
                          <TrendingDown class="h-3 w-3 text-red-500" />
                          <span class="text-xs text-red-500">
                            {changeData.percentage.toFixed(1)}%
                          </span>
                        {/if}
                      </div>
                    {/if}
                  {:else}
                    <span class="text-muted-foreground">Not tracked</span>
                    <Button
                      size="sm"
                      variant="outline"
                      onclick={() => requestTag(tag.tag, index)}
                      disabled={tag.loading}
                    >
                      {#if tag.loading}
                        <Loader2 class="mr-2 h-3 w-3 animate-spin" />
                        Adding...
                      {:else}
                        Add Tag
                      {/if}
                    </Button>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>

    <DialogFooter>
      <Button variant="outline" onclick={() => (open = false)}>Close</Button>
      <Button onclick={analyzePostText} disabled={loading}>
        {#if loading}
          <Loader2 class="mr-2 h-4 w-4 animate-spin" />
          Analyzing...
        {:else}
          Analyze
        {/if}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
