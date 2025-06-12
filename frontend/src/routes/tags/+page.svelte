<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle
  } from '$lib/components/ui/card';
  import { DateRangePicker } from '$lib/components/ui/date-picker';
  import { getLocalTimeZone, today } from '@internationalized/date';
  import type { DateRange } from 'bits-ui';
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

  let { data } = $props();

  let searchInput = $state(data.search);
  let searchInputElement: HTMLInputElement | undefined;
  let expandedTagId = $state<string | null>(null);
  let searchDebounceTimer: number;

  // Date range state
  let dateRangeValue = $state<DateRange>({
    start: today(getLocalTimeZone()).subtract({ days: 7 }),
    end: today(getLocalTimeZone())
  });

  // Tag changes cache
  let tagChanges = $state<Record<string, { change: number; percentage: number }>>({});

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

      // Save current selection state before navigation
      const activeElement = document.activeElement;
      const selectionStart = searchInputElement?.selectionStart;
      const selectionEnd = searchInputElement?.selectionEnd;

      goto(`?${params}`, { replaceState: true, keepFocus: true }).then(() => {
        // Restore focus and selection if the search input was focused
        if (activeElement === searchInputElement && searchInputElement) {
          searchInputElement.focus();
          if (selectionStart !== undefined && selectionEnd !== undefined) {
            searchInputElement.setSelectionRange(selectionStart, selectionEnd);
          }
        }
      });
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
      start: today(getLocalTimeZone()).subtract({ days }),
      end: today(getLocalTimeZone())
    };
    fetchTagChanges();
  }

  async function fetchTagChanges() {
    if (!dateRangeValue?.start || !dateRangeValue?.end) return;

    // Ensure endDate includes the full day by setting time to end of day
    const endDate = dateRangeValue.end!.toDate(getLocalTimeZone());
    endDate.setHours(23, 59, 59, 999);

    // Fetch changes for all visible tags
    const promises = data.tags.map(async (tag) => {
      try {
        const response = await fetch(
          `http://localhost:3000/api/tags/${tag.id}/history?startDate=${dateRangeValue.start!.toDate(getLocalTimeZone()).toISOString()}&endDate=${endDate.toISOString()}`
        );
        if (response.ok) {
          const { history } = (await response.json()) as { history: { viewCount: number }[] };
          if (history.length > 0) {
            // Calculate change by comparing first and last viewCount in the range
            // History is returned in descending order (newest first)
            const newestViewCount = history[0].viewCount;
            const oldestViewCount = history[history.length - 1].viewCount;
            const totalChange = newestViewCount - oldestViewCount;
            const percentage = oldestViewCount > 0 ? (totalChange / oldestViewCount) * 100 : 0;
            tagChanges[tag.id] = { change: totalChange, percentage };
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
          <input
            bind:this={searchInputElement}
            type="search"
            placeholder="Search tags..."
            value={searchInput}
            oninput={(e) => handleSearch(e.currentTarget.value)}
            class="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex h-10 w-full rounded-md border px-3 py-2 pl-9 text-sm file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
          />
        </div>

        <div class="flex gap-2">
          <DateRangePicker
            bind:value={dateRangeValue}
            presets={dateRangePresets}
            onPresetSelect={setDateRangePreset}
          />
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
              {@const changeData = tagChanges[tag.id] || { change: 0, percentage: 0 }}
              {@const change = changeData.change}
              {@const percentage = changeData.percentage}
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
                        <span class="text-green-500">
                          +{formatNumber(change)} ({percentage >= 0 ? '+' : ''}{percentage.toFixed(
                            2
                          )}%)
                        </span>
                      {:else}
                        <TrendingDown class="h-4 w-4 text-red-500" />
                        <span class="text-red-500">
                          {formatNumber(change)} ({percentage.toFixed(2)}%)
                        </span>
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
