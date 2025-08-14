<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import * as Alert from '$lib/components/ui/alert';
  import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle
  } from '$lib/components/ui/card';
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
    ArrowDown,
    ArrowUp,
    ArrowUpDown,
    Search,
    AlertTriangle,
    Ban,
    ChevronLeft,
    ChevronRight,
    ChevronsLeft,
    ChevronsRight
  } from 'lucide-svelte';

  let { data } = $props();

  let searchInputElement: HTMLInputElement | undefined;
  let searchValue = $state(data.search || '');
  let pageJumpValue = $state<string>('');
  let showPageJump = $state(false);

  function handleSort(field: string) {
    const params = new URLSearchParams($page.url.searchParams);

    if (data.sortBy === field) {
      params.set('sortOrder', data.sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      params.set('sortBy', field);
      params.set('sortOrder', field === 'deletedDetectedAt' ? 'desc' : 'asc');
    }

    params.set('page', '1');
    goto(`?${params.toString()}`);
  }

  function handleSearch() {
    const params = new URLSearchParams($page.url.searchParams);
    params.set('search', searchValue);
    params.set('page', '1');
    goto(`?${params.toString()}`);
  }

  function handlePageChange(newPage: number) {
    const params = new URLSearchParams($page.url.searchParams);
    params.set('page', newPage.toString());
    goto(`?${params.toString()}`);
  }

  function goToFirstPage() {
    handlePageChange(1);
  }

  function goToLastPage() {
    handlePageChange(data.pagination.totalPages);
  }

  function handlePageJump() {
    const pageNum = parseInt(pageJumpValue);
    if (!isNaN(pageNum) && pageNum >= 1 && pageNum <= data.pagination.totalPages) {
      handlePageChange(pageNum);
      pageJumpValue = '';
      showPageJump = false;
    }
  }

  function formatNumber(num: number): string {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  }

  function formatDate(timestamp: number | null | undefined): string {
    if (!timestamp) return 'Unknown';
    const date = new Date(timestamp * 1000);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function daysSinceBan(timestamp: number | null | undefined): number {
    if (!timestamp) return 0;
    const banDate = new Date(timestamp * 1000);
    const now = new Date();
    const diffTime = Math.abs(now.getTime() - banDate.getTime());
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays;
  }

  function isRecentlyBanned(timestamp: number | null | undefined): boolean {
    if (!timestamp) return false;
    return daysSinceBan(timestamp) <= 7;
  }

  // Keyboard shortcuts
  $effect(() => {
    const handleKeydown = (e: KeyboardEvent) => {
      if (e.ctrlKey || e.metaKey) {
        if (e.key === 'k' || e.key === 'f') {
          e.preventDefault();
          searchInputElement?.focus();
        }
      }
    };

    window.addEventListener('keydown', handleKeydown);
    return () => window.removeEventListener('keydown', handleKeydown);
  });
</script>

<div class="container mx-auto px-4 py-8">
  <Card>
    <CardHeader>
      <CardTitle class="flex items-center gap-2">
        <Ban class="text-destructive h-5 w-5" />
        Banned Tags
      </CardTitle>
      <CardDescription>Tags that have been removed or banned from Fansly</CardDescription>
    </CardHeader>
    <CardContent class="space-y-4">
      <!-- Warning Banner -->
      <Alert.Root variant="destructive">
        <AlertTriangle class="h-4 w-4" />
        <Alert.Title>Banned Content</Alert.Title>
        <Alert.Description>
          These tags have been detected as deleted or banned from Fansly. They may have been removed
          due to policy violations, content guidelines, or other reasons.
        </Alert.Description>
      </Alert.Root>

      <!-- Statistics Cards -->
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader class="pb-2">
            <CardDescription>Total Banned</CardDescription>
          </CardHeader>
          <CardContent>
            <p class="text-2xl font-bold">{data.statistics.totalBanned}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="pb-2">
            <CardDescription>Last 24 Hours</CardDescription>
          </CardHeader>
          <CardContent>
            <p class="text-destructive text-2xl font-bold">{data.statistics.bannedLast24h}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="pb-2">
            <CardDescription>Last 7 Days</CardDescription>
          </CardHeader>
          <CardContent>
            <p class="text-2xl font-bold text-orange-500">{data.statistics.bannedLast7d}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="pb-2">
            <CardDescription>Last 30 Days</CardDescription>
          </CardHeader>
          <CardContent>
            <p class="text-2xl font-bold">{data.statistics.bannedLast30d}</p>
          </CardContent>
        </Card>
      </div>

      <!-- Search Controls -->
      <div class="flex gap-2">
        <div class="relative flex-1">
          <Search class="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
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
                      #{tag.tag}
                      {#if isRecentlyBanned(tag.deletedDetectedAt)}
                        <Badge variant="destructive" class="text-xs">New</Badge>
                      {/if}
                    </div>
                  </TableCell>
                  <TableCell>
                    <span
                      class={isRecentlyBanned(tag.deletedDetectedAt)
                        ? 'text-destructive font-semibold'
                        : ''}
                    >
                      {formatDate(tag.deletedDetectedAt)}
                    </span>
                  </TableCell>
                  <TableCell>
                    {daysSinceBan(tag.deletedDetectedAt)} days
                  </TableCell>
                  <TableCell>{formatNumber(tag.viewCount)}</TableCell>
                  <TableCell>{formatNumber(tag.postCount)}</TableCell>
                </TableRow>
              {/each}
            </TableBody>
          </Table>
        </div>

        <!-- Pagination -->
        {#if data.pagination.totalPages > 1}
          <div class="flex items-center justify-between">
            <p class="text-muted-foreground text-sm">
              Showing {(data.pagination.page - 1) * data.pagination.limit + 1} -
              {Math.min(data.pagination.page * data.pagination.limit, data.pagination.totalCount)} of
              {data.pagination.totalCount} banned tags
            </p>
            <Pagination.Root
              count={data.pagination.totalCount}
              perPage={data.pagination.limit}
              page={data.pagination.page}
              onPageChange={handlePageChange}
            >
              {#snippet children({ pages, currentPage })}
                <Pagination.Content>
                  <Pagination.Item>
                    <Pagination.PrevButton>
                      <ChevronLeft class="h-4 w-4" />
                      <span class="hidden sm:inline">Previous</span>
                    </Pagination.PrevButton>
                  </Pagination.Item>

                  {#if data.pagination.page > 2}
                    <Pagination.Item>
                      <Button variant="outline" size="icon" onclick={goToFirstPage}>
                        <ChevronsLeft class="h-4 w-4" />
                      </Button>
                    </Pagination.Item>
                  {/if}

                  {#each pages as page (page.key)}
                    {#if page.type === 'ellipsis'}
                      <Pagination.Item>
                        <Pagination.Ellipsis />
                      </Pagination.Item>
                    {:else}
                      <Pagination.Item>
                        <Pagination.Link {page} isActive={currentPage === page.value}>
                          {page.value}
                        </Pagination.Link>
                      </Pagination.Item>
                    {/if}
                  {/each}

                  {#if data.pagination.page < data.pagination.totalPages - 1}
                    <Pagination.Item>
                      <Button variant="outline" size="icon" onclick={goToLastPage}>
                        <ChevronsRight class="h-4 w-4" />
                      </Button>
                    </Pagination.Item>
                  {/if}

                  <Pagination.Item>
                    <Pagination.NextButton>
                      <span class="hidden sm:inline">Next</span>
                      <ChevronRight class="h-4 w-4" />
                    </Pagination.NextButton>
                  </Pagination.Item>

                  {#if showPageJump}
                    <Pagination.Item>
                      <input
                        bind:value={pageJumpValue}
                        type="number"
                        min="1"
                        max={data.pagination.totalPages}
                        placeholder="Page..."
                        onkeydown={(e) => {
                          if (e.key === 'Enter') {
                            handlePageJump();
                          } else if (e.key === 'Escape') {
                            showPageJump = false;
                            pageJumpValue = '';
                          }
                        }}
                        onblur={() => {
                          showPageJump = false;
                          pageJumpValue = '';
                        }}
                        class="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring h-9 w-20 rounded-md border px-3 py-1 text-sm focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
                      />
                    </Pagination.Item>
                  {:else}
                    <Pagination.Item>
                      <Button
                        variant="outline"
                        size="sm"
                        onclick={() => {
                          showPageJump = true;
                        }}
                      >
                        Go to page
                      </Button>
                    </Pagination.Item>
                  {/if}
                </Pagination.Content>
              {/snippet}
            </Pagination.Root>
          </div>
        {/if}
      {/if}
    </CardContent>
  </Card>
</div>
