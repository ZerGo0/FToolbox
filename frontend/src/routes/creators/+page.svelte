<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import CreatorHistory from '$lib/components/CreatorHistory.svelte';
  import CreatorRequestDialog from '$lib/components/CreatorRequestDialog.svelte';
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
  import * as Tooltip from '$lib/components/ui/tooltip';
  import { getLocalTimeZone, today } from '@internationalized/date';
  import type { DateRange } from 'bits-ui';
  import {
    AlertCircle,
    ArrowDown,
    ArrowUp,
    ArrowUpDown,
    ChevronDown,
    ChevronLeft,
    ChevronRight,
    ChevronsLeft,
    ChevronsRight,
    ChevronUp,
    Search,
    TrendingDown,
    TrendingUp
  } from 'lucide-svelte';

  let { data } = $props();

  let searchInputElement: HTMLInputElement | undefined;
  let expandedCreatorId = $state<string | null>(null);
  let searchValue = $state(data.search || '');
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

  function handleSearch() {
    const params = new URLSearchParams($page.url.searchParams);
    params.set('search', searchValue);
    params.set('page', '1');
    goto(`?${params}`);
  }

  function handleSort(column: string) {
    const params = new URLSearchParams($page.url.searchParams);
    const currentSortBy = params.get('sortBy') || 'rank';
    const currentSortOrder = params.get('sortOrder') || 'asc';

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

  function toggleCreator(creatorId: string) {
    expandedCreatorId = expandedCreatorId === creatorId ? null : creatorId;
  }

  function formatNumber(num: number): string {
    return new Intl.NumberFormat().format(num);
  }

  function formatDate(date: Date | string | number | null | undefined): string {
    if (!date) return 'Never';
    // Handle Unix timestamps (numbers)
    // If the number is less than 10 billion, it's likely in seconds, otherwise milliseconds
    const dateObj =
      typeof date === 'number'
        ? date < 1e10
          ? new Date(date * 1000)
          : new Date(date)
        : new Date(date);
    return dateObj.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
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

  // Format creator display name
  function getCreatorDisplayName(creator: { displayName?: string | null; username: string }) {
    if (creator.displayName && creator.displayName !== creator.username) {
      return `${creator.displayName} (@${creator.username})`;
    }
    return `@${creator.username}`;
  }
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <h1 class="text-3xl font-bold">Creators</h1>
    <CreatorRequestDialog />
  </div>

  <div
    class="flex items-start gap-2 rounded-lg border border-amber-200 bg-amber-50 p-4 dark:border-amber-900 dark:bg-amber-950/20"
  >
    <AlertCircle class="mt-0.5 h-5 w-5 text-amber-600 dark:text-amber-400" />
    <div class="text-sm text-amber-800 dark:text-amber-200">
      <p class="font-medium">Data Notice</p>
      <p class="mt-1">
        This data is incomplete and continuously growing. Some creators may be missing as we
        discover and add new ones regularly.
      </p>
    </div>
  </div>

  {#if data.statistics}
    <Card class="mb-6">
      <CardHeader>
        <CardTitle>Global Creator Statistics</CardTitle>
        <CardDescription>Platform-wide creator performance metrics</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="grid gap-4 sm:grid-cols-3">
          <!-- Followers -->
          <div class="space-y-2">
            <p class="text-muted-foreground text-sm">Total Followers</p>
            <p class="text-2xl font-bold sm:text-3xl">
              {formatNumber(data.statistics.totalFollowers)}
            </p>
            <p class="text-muted-foreground text-sm">24-hour Change</p>
            {#if data.statistics.followersChange24h !== 0}
              <div class="flex items-center gap-1">
                {#if data.statistics.followersChange24h > 0}
                  <TrendingUp class="h-4 w-4 text-green-500" />
                  <span class="text-sm font-semibold text-green-500">
                    +{formatNumber(data.statistics.followersChange24h)}
                  </span>
                {:else}
                  <TrendingDown class="h-4 w-4 text-red-500" />
                  <span class="text-sm font-semibold text-red-500">
                    {formatNumber(data.statistics.followersChange24h)}
                  </span>
                {/if}
                <span class="text-muted-foreground text-sm">
                  ({data.statistics.followersChangePercent24h > 0
                    ? '+'
                    : ''}{data.statistics.followersChangePercent24h.toFixed(2)}%)
                </span>
              </div>
            {:else}
              <p class="text-muted-foreground text-sm">No change</p>
            {/if}
          </div>

          <!-- Media Likes -->
          <div class="space-y-2 sm:text-center">
            <p class="text-muted-foreground text-sm">Total Media Likes</p>
            <p class="text-2xl font-bold sm:text-3xl">
              {formatNumber(data.statistics.totalMediaLikes)}
            </p>
            <p class="text-muted-foreground text-sm">24-hour Change</p>
            {#if data.statistics.mediaLikesChange24h !== 0}
              <div class="flex items-center gap-1 sm:justify-center">
                {#if data.statistics.mediaLikesChange24h > 0}
                  <TrendingUp class="h-4 w-4 text-green-500" />
                  <span class="text-sm font-semibold text-green-500">
                    +{formatNumber(data.statistics.mediaLikesChange24h)}
                  </span>
                {:else}
                  <TrendingDown class="h-4 w-4 text-red-500" />
                  <span class="text-sm font-semibold text-red-500">
                    {formatNumber(data.statistics.mediaLikesChange24h)}
                  </span>
                {/if}
                <span class="text-muted-foreground text-sm">
                  ({data.statistics.mediaLikesChangePercent24h > 0
                    ? '+'
                    : ''}{data.statistics.mediaLikesChangePercent24h.toFixed(2)}%)
                </span>
              </div>
            {:else}
              <p class="text-muted-foreground text-sm">No change</p>
            {/if}
          </div>

          <!-- Post Likes -->
          <div class="space-y-2 sm:text-right">
            <p class="text-muted-foreground text-sm">Total Post Likes</p>
            <p class="text-2xl font-bold sm:text-3xl">
              {formatNumber(data.statistics.totalPostLikes)}
            </p>
            <p class="text-muted-foreground text-sm">24-hour Change</p>
            {#if data.statistics.postLikesChange24h !== 0}
              <div class="flex items-center gap-1 sm:justify-end">
                {#if data.statistics.postLikesChange24h > 0}
                  <TrendingUp class="h-4 w-4 text-green-500" />
                  <span class="text-sm font-semibold text-green-500">
                    +{formatNumber(data.statistics.postLikesChange24h)}
                  </span>
                {:else}
                  <TrendingDown class="h-4 w-4 text-red-500" />
                  <span class="text-sm font-semibold text-red-500">
                    {formatNumber(data.statistics.postLikesChange24h)}
                  </span>
                {/if}
                <span class="text-muted-foreground text-sm">
                  ({data.statistics.postLikesChangePercent24h > 0
                    ? '+'
                    : ''}{data.statistics.postLikesChangePercent24h.toFixed(2)}%)
                </span>
              </div>
            {:else}
              <p class="text-muted-foreground text-sm">No change</p>
            {/if}
          </div>
        </div>
      </CardContent>
    </Card>
  {/if}

  <Card>
    <CardHeader>
      <CardTitle>Creator Statistics</CardTitle>
      <CardDescription>Monitor and analyze Fansly creator performance over time</CardDescription>
    </CardHeader>
    <CardContent class="space-y-4">
      <!-- Search and Date Range Controls -->
      <div class="flex flex-col gap-4 sm:flex-row">
        <div class="flex flex-1 gap-2">
          <div class="relative flex-1">
            <Search
              class="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2"
            />
            <input
              bind:this={searchInputElement}
              bind:value={searchValue}
              type="search"
              placeholder="Search creators..."
              onkeydown={(e) => {
                if (e.key === 'Enter') {
                  handleSearch();
                }
              }}
              class="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex h-10 w-full rounded-md border px-3 py-2 pl-9 text-sm file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
            />
          </div>
          <Button onclick={handleSearch} variant="default" size="default" class="h-10">
            <Search class="h-4 w-4 md:mr-2" />
            <span class="hidden md:inline">Search</span>
          </Button>
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

      <!-- Creators Table -->
      {#if data.pagination.totalCount === 0}
        <div class="rounded-md border p-8">
          <p class="text-muted-foreground text-center">No creators found</p>
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
                    onclick={() => handleSort('username')}
                  >
                    Creator
                    {#if data.sortBy === 'username'}
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
                    onclick={() => handleSort('followers')}
                  >
                    Followers
                    {#if data.sortBy === 'followers'}
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
                    onclick={() => handleSort('mediaLikes')}
                  >
                    Media Likes
                    {#if data.sortBy === 'mediaLikes'}
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
                    onclick={() => handleSort('postLikes')}
                  >
                    Post Likes
                    {#if data.sortBy === 'postLikes'}
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
                    onclick={() => handleSort('imageCount')}
                  >
                    Image Count
                    {#if data.sortBy === 'imageCount'}
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
                    onclick={() => handleSort('videoCount')}
                  >
                    Video Count
                    {#if data.sortBy === 'videoCount'}
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
              {#each data.creators as creator (creator.id)}
                <TableRow>
                  <TableCell>
                    <button
                      onclick={() => toggleCreator(creator.id)}
                      class="hover:bg-muted rounded p-1 transition-colors"
                    >
                      {#if expandedCreatorId === creator.id}
                        <ChevronUp class="h-4 w-4" />
                      {:else}
                        <ChevronDown class="h-4 w-4" />
                      {/if}
                    </button>
                  </TableCell>
                  <TableCell class="text-center font-medium">{creator.rank ?? '-'}</TableCell>
                  <TableCell class="font-medium">
                    <div class="flex items-center gap-2">
                      <span>{getCreatorDisplayName(creator)}</span>
                      {#if creator.isDeleted}
                        <Tooltip.Root>
                          <Tooltip.Trigger>
                            <Badge variant="destructive" class="p-1">
                              <AlertCircle class="h-3 w-3" />
                            </Badge>
                          </Tooltip.Trigger>
                          <Tooltip.Content>
                            No longer exists on Fansly (detected {formatDate(
                              creator.deletedDetectedAt
                            )})
                          </Tooltip.Content>
                        </Tooltip.Root>
                      {/if}
                    </div>
                  </TableCell>
                  <TableCell>{formatNumber(creator.followers)}</TableCell>
                  <TableCell>{formatNumber(creator.mediaLikes)}</TableCell>
                  <TableCell>{formatNumber(creator.postLikes)}</TableCell>
                  <TableCell>{formatNumber(creator.imageCount)}</TableCell>
                  <TableCell>{formatNumber(creator.videoCount)}</TableCell>
                  <TableCell>{formatDate(creator.lastCheckedAt)}</TableCell>
                </TableRow>
                {#if expandedCreatorId === creator.id}
                  <TableRow>
                    <TableCell colspan={9} class="p-0">
                      <div class="bg-muted/50 p-6">
                        <CreatorHistory history={creator.history} />
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
