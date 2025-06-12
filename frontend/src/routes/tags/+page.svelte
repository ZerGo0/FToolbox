<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import TagHistory from '$lib/components/TagHistory.svelte';
  import TagRequestDialog from '$lib/components/TagRequestDialog.svelte';
  import PostTextDialog from '$lib/components/PostTextDialog.svelte';
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
  import { Input } from '$lib/components/ui/input';
  import * as Pagination from '$lib/components/ui/pagination';
  import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow
  } from '$lib/components/ui/table';
  import { getLocalTimeZone, today } from '@internationalized/date';
  import type { DateRange } from 'bits-ui';
  import {
    ArrowDown,
    ArrowUp,
    ArrowUpDown,
    ChevronDown,
    ChevronUp,
    Search,
    TrendingDown,
    TrendingUp,
    ChevronLeft,
    ChevronRight,
    ChevronsLeft,
    ChevronsRight
  } from 'lucide-svelte';

  let { data } = $props();

  let searchInputElement: HTMLInputElement | undefined;
  let expandedTagId = $state<string | null>(null);
  let searchDebounceTimer: number;
  let pageJumpValue = $state<string>('');
  let showPageJump = $state(false);

  // Date range state
  let dateRangeValue = $state<DateRange>({
    start: today(getLocalTimeZone()).subtract({ days: 7 }),
    end: today(getLocalTimeZone())
  });

  // Initialize date range from URL params
  $effect(() => {
    if (data.historyStartDate && data.historyEndDate) {
      try {
        const start = new Date(data.historyStartDate);
        const end = new Date(data.historyEndDate);
        dateRangeValue = {
          start: today(getLocalTimeZone()).set({
            year: start.getFullYear(),
            month: start.getMonth() + 1,
            day: start.getDate()
          }),
          end: today(getLocalTimeZone()).set({
            year: end.getFullYear(),
            month: end.getMonth() + 1,
            day: end.getDate()
          })
        };
      } catch {
        // Invalid dates, keep defaults
      }
    }
  });

  const dateRangePresets = [
    { label: '1 Day', days: 1 },
    { label: '1 Week', days: 7 },
    { label: '1 Month', days: 30 },
    { label: '3 Months', days: 90 },
    { label: '6 Months', days: 180 },
    { label: '1 Year', days: 365 }
  ];

  function handleSearch(value: string) {
    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = setTimeout(() => {
      const params = new URLSearchParams($page.url.searchParams);
      params.set('search', value);
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

  function handlePageJump() {
    const pageNumber = parseInt(pageJumpValue);
    if (!isNaN(pageNumber) && pageNumber >= 1 && pageNumber <= data.pagination.totalPages) {
      handlePageChange(pageNumber);
      pageJumpValue = '';
    }
  }

  function toggleTag(tagId: string) {
    expandedTagId = expandedTagId === tagId ? null : tagId;
  }

  function formatNumber(num: number): string {
    return new Intl.NumberFormat().format(num);
  }

  function formatDate(date: Date | string | number | null): string {
    if (!date) return 'Never';
    // Handle Unix timestamps (numbers)
    const dateObj = typeof date === 'number' ? new Date(date * 1000) : new Date(date);
    return dateObj.toLocaleDateString('en-US', {
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
    // Don't automatically apply - user needs to click Apply button
  }

  function updateDateRangeParams() {
    if (!dateRangeValue?.start || !dateRangeValue?.end) return;

    const params = new URLSearchParams($page.url.searchParams);

    // Ensure endDate includes the full day by setting time to end of day
    const startDate = dateRangeValue.start.toDate(getLocalTimeZone());
    const endDate = dateRangeValue.end.toDate(getLocalTimeZone());
    endDate.setHours(23, 59, 59, 999);

    params.set('historyStartDate', startDate.toISOString());
    params.set('historyEndDate', endDate.toISOString());
    params.set('includeHistory', 'true');

    goto(`?${params}`);
  }

  // Manual update button for date range
  function applyDateRange() {
    updateDateRangeParams();
  }

  // Calculate change from history data
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
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <h1 class="text-3xl font-bold">Tags</h1>
    <div class="flex gap-2">
      <PostTextDialog />
      <TagRequestDialog />
    </div>
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
            value={data.search}
            oninput={(e) => handleSearch(e.currentTarget.value)}
            class="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex h-10 w-full rounded-md border px-3 py-2 pl-9 text-sm file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
          />
        </div>

        <div class="flex gap-2">
          <DateRangePicker
            bind:value={dateRangeValue}
            presets={dateRangePresets}
            onPresetSelect={setDateRangePreset}
            onApply={applyDateRange}
          />
        </div>
      </div>

      <!-- Tags Table -->
      {#if data.pagination.totalCount === 0}
        <div class="rounded-md border p-8">
          <p class="text-muted-foreground text-center">No tags found</p>
        </div>
      {:else}
        <div class="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="w-12"></TableHead>
                <TableHead class="w-16">
                  <button
                    class="hover:text-foreground flex items-center gap-1 transition-colors"
                    onclick={() => handleSort('rank')}
                  >
                    Rank
                    {#if data.sortBy === 'rank'}
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
                <TableHead>
                  <button
                    class="hover:text-foreground flex items-center gap-1 transition-colors"
                    onclick={() => handleSort('change')}
                  >
                    Change
                    {#if data.sortBy === 'change'}
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
                {@const changeData = getTagChange(tag)}
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
                            +{formatNumber(change)} ({percentage >= 0
                              ? '+'
                              : ''}{percentage.toFixed(2)}%)
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
                        <TagHistory history={tag.history} />
                      </div>
                    </TableCell>
                  </TableRow>
                {/if}
              {/each}
            </TableBody>
          </Table>
        </div>
      {/if}

      <!-- Pagination -->
      {#if data.pagination.totalPages > 1}
        <div class="flex justify-center">
          <Pagination.Root
            count={data.pagination.totalCount}
            perPage={data.pagination.limit}
            page={data.pagination.page}
            onPageChange={(page) => handlePageChange(page)}
            siblingCount={1}
          >
            {#snippet children({ pages, currentPage })}
              <Pagination.Content class="flex items-center gap-1">
                <!-- First Page -->
                <Pagination.Item class="hidden sm:block">
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-9 w-9"
                    disabled={currentPage <= 1}
                    onclick={() => handlePageChange(1)}
                    aria-label="Go to first page"
                  >
                    <ChevronsLeft class="h-4 w-4" />
                  </Button>
                </Pagination.Item>

                <!-- Previous -->
                <Pagination.Item>
                  <Pagination.PrevButton class="h-9 px-3">
                    <ChevronLeft class="h-4 w-4" />
                    <span class="sr-only sm:not-sr-only sm:ml-1">Previous</span>
                  </Pagination.PrevButton>
                </Pagination.Item>

                <!-- Page Numbers - hide some on mobile -->
                {#each pages as page (page.key)}
                  {#if page.type === 'ellipsis'}
                    <Pagination.Item>
                      {#if showPageJump}
                        <div class="flex items-center gap-1">
                          <Input
                            type="number"
                            class="h-9 w-16"
                            placeholder="..."
                            bind:value={pageJumpValue}
                            onkeydown={(e) => {
                              if (e.key === 'Enter') {
                                handlePageJump();
                                showPageJump = false;
                              } else if (e.key === 'Escape') {
                                showPageJump = false;
                                pageJumpValue = '';
                              }
                            }}
                            onblur={() => {
                              setTimeout(() => {
                                showPageJump = false;
                                pageJumpValue = '';
                              }, 200);
                            }}
                            min="1"
                            max={data.pagination.totalPages}
                            autofocus
                          />
                          <Button
                            size="icon"
                            variant="ghost"
                            class="h-9 w-9"
                            onclick={() => {
                              handlePageJump();
                              showPageJump = false;
                            }}
                          >
                            <ChevronRight class="h-4 w-4" />
                          </Button>
                        </div>
                      {:else}
                        <Button
                          variant="ghost"
                          size="icon"
                          class="h-9 w-9"
                          onclick={() => {
                            showPageJump = true;
                            pageJumpValue = '';
                          }}
                          title="Go to page..."
                        >
                          <span class="text-muted-foreground">...</span>
                        </Button>
                      {/if}
                    </Pagination.Item>
                  {:else}
                    <!-- On mobile, only show current page and adjacent pages -->
                    <Pagination.Item
                      class={page.value !== currentPage &&
                      Math.abs(page.value - currentPage) > 1 &&
                      page.value !== 1 &&
                      page.value !== data.pagination.totalPages
                        ? 'hidden sm:block'
                        : ''}
                    >
                      <Pagination.Link {page} isActive={currentPage === page.value} class="h-9 w-9">
                        {page.value}
                      </Pagination.Link>
                    </Pagination.Item>
                  {/if}
                {/each}

                <!-- Next -->
                <Pagination.Item>
                  <Pagination.NextButton class="h-9 px-3">
                    <span class="sr-only sm:not-sr-only sm:mr-1">Next</span>
                    <ChevronRight class="h-4 w-4" />
                  </Pagination.NextButton>
                </Pagination.Item>

                <!-- Last Page -->
                <Pagination.Item class="hidden sm:block">
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-9 w-9"
                    disabled={currentPage >= data.pagination.totalPages}
                    onclick={() => handlePageChange(data.pagination.totalPages)}
                    aria-label="Go to last page"
                  >
                    <ChevronsRight class="h-4 w-4" />
                  </Button>
                </Pagination.Item>
              </Pagination.Content>
            {/snippet}
          </Pagination.Root>
        </div>
      {/if}
    </CardContent>
  </Card>
</div>
