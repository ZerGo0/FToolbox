<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle
  } from '$lib/components/ui/card';
  import { Popover, PopoverContent, PopoverTrigger } from '$lib/components/ui/popover';
  import {
    Table,
    TableBody,
    TableCaption,
    TableCell,
    TableHead,
    TableHeader,
    TableRow
  } from '$lib/components/ui/table';
  import {
    CalendarIcon,
    ChevronDown,
    ChevronUp,
    ArrowUpDown,
    ArrowUp,
    ArrowDown,
    Search,
    TrendingUp,
    TrendingDown
  } from 'lucide-svelte';
  import TagHistory from '$lib/components/TagHistory.svelte';
  import TagRequestDialog from '$lib/components/TagRequestDialog.svelte';
  import { format, subDays, startOfDay, endOfDay } from 'date-fns';

  let { data } = $props();

  let searchInput = $state(data.search);
  let expandedTagId = $state<string | null>(null);
  let searchDebounceTimer: number;

  // Date range state
  let dateRangeValue = $state<{ start: Date; end: Date } | undefined>({
    start: subDays(new Date(), 7),
    end: new Date()
  });

  // Tag changes cache
  let tagChanges = $state<Record<string, number>>({});

  const dateRangePresets = [
    { label: '1 Day', days: 1 },
    { label: '1 Week', days: 7 },
    { label: '1 Month', days: 30 },
    { label: '3 Months', days: 90 },
    { label: '6 Months', days: 180 },
    { label: '1 Year', days: 365 }
  ];

  function handleSearch(value: string) {
    searchInput = value;
    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = setTimeout(() => {
      const params = new URLSearchParams($page.url.searchParams);
      params.set('search', searchInput);
      params.set('page', '1');
      goto(`?${params}`);
    }, 300);
  }

  function handleSort(column: string) {
    const params = new URLSearchParams($page.url.searchParams);
    const currentSortBy = params.get('sortBy') || 'viewCount';
    const currentSortOrder = params.get('sortOrder') || 'desc';

    if (currentSortBy === column) {
      params.set('sortOrder', currentSortOrder === 'desc' ? 'asc' : 'desc');
    } else {
      params.set('sortBy', column);
      params.set('sortOrder', 'desc');
    }
    params.set('page', '1');
    goto(`?${params}`);
  }

  function handlePageChange(newPage: number) {
    const params = new URLSearchParams($page.url.searchParams);
    params.set('page', newPage.toString());
    goto(`?${params}`);
  }

  function toggleTag(tagId: string) {
    expandedTagId = expandedTagId === tagId ? null : tagId;
  }

  function formatNumber(num: number): string {
    return new Intl.NumberFormat().format(num);
  }

  function formatDate(date: Date | string | null): string {
    if (!date) return 'Never';
    return new Date(date).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function setDateRangePreset(days: number) {
    dateRangeValue = {
      start: startOfDay(subDays(new Date(), days)),
      end: endOfDay(new Date())
    };
    fetchTagChanges();
  }

  async function fetchTagChanges() {
    if (!dateRangeValue?.start || !dateRangeValue?.end) return;

    // Fetch changes for all visible tags
    const promises = data.tags.map(async (tag) => {
      try {
        const response = await fetch(
          `http://localhost:3000/api/tags/${tag.id}/history?startDate=${dateRangeValue!.start.toISOString()}&endDate=${dateRangeValue!.end.toISOString()}`
        );
        if (response.ok) {
          const { history } = (await response.json()) as { history: { change: number }[] };
          if (history.length > 0) {
            const totalChange = history.reduce((sum: number, h) => sum + h.change, 0);
            tagChanges[tag.id] = totalChange;
          }
        }
      } catch (error) {
        console.error(`Failed to fetch changes for tag ${tag.id}`, error);
      }
    });

    await Promise.all(promises);
  }

  // Fetch changes when component mounts or date range changes
  $effect(() => {
    if (dateRangeValue?.start && dateRangeValue?.end) {
      fetchTagChanges();
    }
  });
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <h1 class="text-3xl font-bold">Tags</h1>
    <TagRequestDialog />
  </div>

  <Card>
    <CardHeader>
      <CardTitle>Tag Statistics</CardTitle>
      <CardDescription>Monitor and analyze Fansly tag performance over time</CardDescription>
    </CardHeader>
    <CardContent class="space-y-4">
      <!-- Search and Date Range Controls -->
      <div class="flex flex-col gap-4 sm:flex-row">
        <div class="relative flex-1">
          <Search class="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
          <Input
            type="search"
            placeholder="Search tags..."
            value={searchInput}
            oninput={(e) => handleSearch(e.currentTarget.value)}
            class="pl-9"
          />
        </div>

        <div class="flex gap-2">
          <Popover>
            <PopoverTrigger>
              <Button variant="outline" class="justify-start text-left font-normal">
                <CalendarIcon class="mr-2 h-4 w-4" />
                {#if dateRangeValue?.start && dateRangeValue?.end}
                  {format(dateRangeValue.start, 'MMM dd')} - {format(
                    dateRangeValue.end,
                    'MMM dd, yyyy'
                  )}
                {:else}
                  Select date range
                {/if}
              </Button>
            </PopoverTrigger>
            <PopoverContent class="w-auto p-0" align="end">
              <div class="space-y-2 p-3">
                <div class="mb-2 text-sm font-medium">Quick select</div>
                <div class="grid grid-cols-2 gap-2">
                  {#each dateRangePresets as preset (preset.days)}
                    <Button
                      variant="outline"
                      size="sm"
                      onclick={() => setDateRangePreset(preset.days)}
                    >
                      {preset.label}
                    </Button>
                  {/each}
                </div>
              </div>
              <div class="p-3">
                <div class="text-muted-foreground text-sm">
                  Custom date range selection coming soon
                </div>
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </div>

      <!-- Tags Table -->
      <div class="rounded-md border">
        <Table>
          <TableCaption>
            Showing {data.tags.length} of {data.pagination.totalCount} tags
          </TableCaption>
          <TableHeader>
            <TableRow>
              <TableHead class="w-12"></TableHead>
              <TableHead class="w-16">Rank</TableHead>
              <TableHead>
                <button
                  class="hover:text-foreground flex items-center gap-1 transition-colors"
                  onclick={() => handleSort('tag')}
                >
                  Tag
                  {#if data.sortBy === 'tag'}
                    {#if data.sortOrder === 'desc'}
                      <ArrowDown class="h-4 w-4" />
                    {:else}
                      <ArrowUp class="h-4 w-4" />
                    {/if}
                  {:else}
                    <ArrowUpDown class="h-4 w-4" />
                  {/if}
                </button>
              </TableHead>
              <TableHead>
                <button
                  class="hover:text-foreground flex items-center gap-1 transition-colors"
                  onclick={() => handleSort('viewCount')}
                >
                  View Count
                  {#if data.sortBy === 'viewCount'}
                    {#if data.sortOrder === 'desc'}
                      <ArrowDown class="h-4 w-4" />
                    {:else}
                      <ArrowUp class="h-4 w-4" />
                    {/if}
                  {:else}
                    <ArrowUpDown class="h-4 w-4" />
                  {/if}
                </button>
              </TableHead>
              <TableHead>Change</TableHead>
              <TableHead>
                <button
                  class="hover:text-foreground flex items-center gap-1 transition-colors"
                  onclick={() => handleSort('updatedAt')}
                >
                  Last Updated
                  {#if data.sortBy === 'updatedAt'}
                    {#if data.sortOrder === 'desc'}
                      <ArrowDown class="h-4 w-4" />
                    {:else}
                      <ArrowUp class="h-4 w-4" />
                    {/if}
                  {:else}
                    <ArrowUpDown class="h-4 w-4" />
                  {/if}
                </button>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {#each data.tags as tag (tag.id)}
              {@const change = tagChanges[tag.id] || 0}
              <TableRow>
                <TableCell>
                  <button
                    onclick={() => toggleTag(tag.id)}
                    class="hover:bg-muted rounded p-1 transition-colors"
                  >
                    {#if expandedTagId === tag.id}
                      <ChevronUp class="h-4 w-4" />
                    {:else}
                      <ChevronDown class="h-4 w-4" />
                    {/if}
                  </button>
                </TableCell>
                <TableCell class="text-center font-medium">{tag.rank ?? '-'}</TableCell>
                <TableCell class="font-medium">
                  <Badge variant="secondary">#{tag.tag}</Badge>
                </TableCell>
                <TableCell>{formatNumber(tag.viewCount)}</TableCell>
                <TableCell>
                  {#if change !== 0}
                    <div class="flex items-center gap-1">
                      {#if change > 0}
                        <TrendingUp class="h-4 w-4 text-green-500" />
                        <span class="text-green-500">+{formatNumber(change)}</span>
                      {:else}
                        <TrendingDown class="h-4 w-4 text-red-500" />
                        <span class="text-red-500">{formatNumber(change)}</span>
                      {/if}
                    </div>
                  {:else}
                    <span class="text-muted-foreground">-</span>
                  {/if}
                </TableCell>
                <TableCell>{formatDate(tag.lastCheckedAt)}</TableCell>
              </TableRow>
              {#if expandedTagId === tag.id}
                <TableRow>
                  <TableCell colspan={6} class="p-0">
                    <div class="bg-muted/50 p-6">
                      <TagHistory tagId={tag.id} dateRange={dateRangeValue} />
                    </div>
                  </TableCell>
                </TableRow>
              {/if}
            {/each}
          </TableBody>
        </Table>
      </div>

      <!-- Pagination -->
      {#if data.pagination.totalPages > 1}
        <div class="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={data.pagination.page <= 1}
            onclick={() => handlePageChange(data.pagination.page - 1)}
          >
            Previous
          </Button>
          <span class="px-4 text-sm">
            Page {data.pagination.page} of {data.pagination.totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={data.pagination.page >= data.pagination.totalPages}
            onclick={() => handlePageChange(data.pagination.page + 1)}
          >
            Next
          </Button>
        </div>
      {/if}
    </CardContent>
  </Card>
</div>
