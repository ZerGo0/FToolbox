<script lang="ts">
  import { goto } from '$app/navigation';
  import { navigating, page } from '$app/stores';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle
  } from '$lib/components/ui/card';
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
  import {
    AlertCircle,
    ArrowDown,
    ArrowUp,
    ArrowUpDown,
    ChevronLeft,
    ChevronRight,
    ChevronsLeft,
    ChevronsRight,
    Search
  } from 'lucide-svelte';
  import { toast } from 'svelte-sonner';

  let { data } = $props();

  let searchInputElement: HTMLInputElement | undefined;
  let searchValue = $state(data.search || '');
  let pageJumpValue = $state<string>('');
  let showPageJump = $state(false);
  let loadingTimeout: ReturnType<typeof setTimeout> | undefined;
  let loadingToastId: string | number | undefined;

  // Show loading toast after 1 second of navigation
  $effect(() => {
    if ($navigating) {
      // Start a timer to show loading toast after 1 second
      loadingTimeout = setTimeout(() => {
        loadingToastId = toast.loading('Loading banned tags...');
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

  function handleSearch() {
    const params = new URLSearchParams($page.url.searchParams);
    params.set('search', searchValue);
    params.set('page', '1');
    goto(`?${params}`);
  }

  function handleSort(column: string) {
    const params = new URLSearchParams($page.url.searchParams);
    const currentSortBy = params.get('sortBy') || 'deletedDetectedAt';
    const currentSortOrder = params.get('sortOrder') || 'desc';

    if (currentSortBy === column) {
      params.set('sortOrder', currentSortOrder === 'desc' ? 'asc' : 'desc');
    } else {
      params.set('sortBy', column);
      params.set('sortOrder', column === 'deletedDetectedAt' ? 'desc' : 'asc');
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

  function formatNumber(num: number): string {
    return new Intl.NumberFormat().format(num);
  }

  function formatDate(timestamp: number | null | undefined): string {
    if (!timestamp) return 'Never';
    // Handle Unix timestamps
    const dateObj = timestamp < 1e10 ? new Date(timestamp * 1000) : new Date(timestamp);
    return dateObj.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function daysSinceBan(timestamp: number | null | undefined): number {
    if (!timestamp) return 0;
    const banDate = new Date(timestamp < 1e10 ? timestamp * 1000 : timestamp);
    const now = new Date();
    const diffTime = Math.abs(now.getTime() - banDate.getTime());
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays;
  }
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <h1 class="text-3xl font-bold">Banned Tags</h1>
  </div>

  <div
    class="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-900 dark:bg-red-950/20"
  >
    <AlertCircle class="mt-0.5 h-5 w-5 text-red-600 dark:text-red-400" />
    <div class="text-sm text-red-800 dark:text-red-200">
      <p class="font-medium">Banned Content</p>
      <p class="mt-1">
        These tags have been detected as deleted or banned from Fansly. They may have been removed
        due to policy violations, content guidelines, or other reasons. Some may have been false
        positives and are not actually deleted/banned.
      </p>
    </div>
  </div>

  {#if data.statistics}
    <Card class="mb-6">
      <CardHeader>
        <CardTitle>Banned Tag Statistics</CardTitle>
        <CardDescription>Overview of banned and removed tags</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="space-y-2">
            <p class="text-muted-foreground text-sm">Total Banned</p>
            <p class="text-2xl font-bold sm:text-3xl">
              {formatNumber(data.statistics.totalBanned || 0)}
            </p>
          </div>
          <div class="space-y-2">
            <p class="text-muted-foreground text-sm">Last 24 Hours</p>
            <p class="text-2xl font-bold text-red-500 sm:text-3xl">
              {formatNumber(data.statistics.bannedLast24h || 0)}
            </p>
          </div>
          <div class="space-y-2">
            <p class="text-muted-foreground text-sm">Last 7 Days</p>
            <p class="text-2xl font-bold text-orange-500 sm:text-3xl">
              {formatNumber(data.statistics.bannedLast7d || 0)}
            </p>
          </div>
          <div class="space-y-2">
            <p class="text-muted-foreground text-sm">Last 30 Days</p>
            <p class="text-2xl font-bold sm:text-3xl">
              {formatNumber(data.statistics.bannedLast30d || 0)}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  {/if}

  <Card>
    <CardHeader>
      <CardTitle>Banned Tags</CardTitle>
      <CardDescription>Tags that have been removed from Fansly</CardDescription>
    </CardHeader>
    <CardContent class="space-y-4">
      <!-- Search Controls -->
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
              placeholder="Search banned tags..."
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
      </div>

      <!-- Tags Table -->
      {#if data.pagination.totalCount === 0}
        <div class="rounded-md border p-8">
          <p class="text-muted-foreground text-center">No banned tags found</p>
        </div>
      {:else}
        <div class="relative rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
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
                    onclick={() => handleSort('deletedDetectedAt')}
                  >
                    Ban Date
                    {#if data.sortBy === 'deletedDetectedAt'}
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
                <TableHead>Days Since Ban</TableHead>
                <TableHead>
                  <button
                    class="hover:text-foreground flex items-center gap-1 transition-colors"
                    onclick={() => handleSort('viewCount')}
                  >
                    Last View Count
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
                    onclick={() => handleSort('postCount')}
                  >
                    Last Post Count
                    {#if data.sortBy === 'postCount'}
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
                <TableRow>
                  <TableCell class="font-medium">
                    <div class="flex items-center gap-2">
                      <Badge variant="destructive">#{tag.tag}</Badge>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span>
                      {formatDate(tag.deletedDetectedAt)}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span class="text-muted-foreground">
                      {daysSinceBan(tag.deletedDetectedAt)} days
                    </span>
                  </TableCell>
                  <TableCell>{formatNumber(tag.viewCount)}</TableCell>
                  <TableCell>{formatNumber(tag.postCount || 0)}</TableCell>
                </TableRow>
              {/each}
            </TableBody>
          </Table>
        </div>

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
                        <Pagination.Link
                          {page}
                          isActive={currentPage === page.value}
                          class="h-9 w-9"
                        >
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
      {/if}
    </CardContent>
  </Card>
</div>
