<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import {
		Table,
		TableBody,
		TableCaption,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import { Collapsible, CollapsibleContent } from '$lib/components/ui/collapsible';
	import { ChevronDown, ChevronUp, ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-svelte';
	import TagHistory from '$lib/components/TagHistory.svelte';
	import TagRequestDialog from '$lib/components/TagRequestDialog.svelte';
	import WorkerStatus from '$lib/components/WorkerStatus.svelte';

	export let data;

	let searchInput = data.search;
	let expandedTagId: string | null = null;

	function handleSearch() {
		const params = new URLSearchParams($page.url.searchParams);
		params.set('search', searchInput);
		params.set('page', '1');
		goto(`?${params}`);
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

	function formatDate(date: Date | string): string {
		return new Date(date).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getSortIcon(column: string) {
		if (data.sortBy !== column) return ArrowUpDown;
		return data.sortOrder === 'desc' ? ArrowDown : ArrowUp;
	}
</script>

<div class="container mx-auto py-8">
	<div class="mb-6 flex items-center justify-between">
		<h1 class="text-3xl font-bold">Fansly Tag Statistics</h1>
		<TagRequestDialog />
	</div>

	<WorkerStatus />

	<div class="mb-6">
		<form
			onsubmit={(e) => {
				e.preventDefault();
				handleSearch();
			}}
			class="flex gap-2"
		>
			<Input type="search" placeholder="Search tags..." bind:value={searchInput} class="max-w-md" />
			<Button type="submit">Search</Button>
		</form>
	</div>

	<Table>
		<TableCaption>
			Showing {data.tags.length} of {data.pagination.totalCount} tags
		</TableCaption>
		<TableHeader>
			<TableRow>
				<TableHead class="w-12"></TableHead>
				<TableHead>
					<button
						class="hover:text-foreground flex items-center gap-1 transition-colors"
						onclick={() => handleSort('tag')}
					>
						Tag
						<svelte:component this={getSortIcon('tag')} class="h-4 w-4" />
					</button>
				</TableHead>
				<TableHead>
					<button
						class="hover:text-foreground flex items-center gap-1 transition-colors"
						onclick={() => handleSort('viewCount')}
					>
						View Count
						<svelte:component this={getSortIcon('viewCount')} class="h-4 w-4" />
					</button>
				</TableHead>
				<TableHead>
					<button
						class="hover:text-foreground flex items-center gap-1 transition-colors"
						onclick={() => handleSort('updatedAt')}
					>
						Last Updated
						<svelte:component this={getSortIcon('updatedAt')} class="h-4 w-4" />
					</button>
				</TableHead>
			</TableRow>
		</TableHeader>
		<TableBody>
			{#each data.tags as tag (tag.id)}
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
					<TableCell class="font-medium">
						<Badge variant="secondary">#{tag.tag}</Badge>
					</TableCell>
					<TableCell>{formatNumber(tag.viewCount)}</TableCell>
					<TableCell>{tag.lastCheckedAt ? formatDate(tag.lastCheckedAt) : 'Never'}</TableCell>
				</TableRow>
				{#if expandedTagId === tag.id}
					<TableRow>
						<TableCell colspan={4} class="p-0">
							<Collapsible open={true}>
								<CollapsibleContent>
									<div class="bg-muted/50 p-6">
										<TagHistory tagId={tag.id} />
									</div>
								</CollapsibleContent>
							</Collapsible>
						</TableCell>
					</TableRow>
				{/if}
			{/each}
		</TableBody>
	</Table>

	{#if data.pagination.totalPages > 1}
		<div class="mt-6 flex justify-center gap-2">
			<Button
				variant="outline"
				disabled={data.pagination.page <= 1}
				onclick={() => handlePageChange(data.pagination.page - 1)}
			>
				Previous
			</Button>
			<span class="flex items-center px-4">
				Page {data.pagination.page} of {data.pagination.totalPages}
			</span>
			<Button
				variant="outline"
				disabled={data.pagination.page >= data.pagination.totalPages}
				onclick={() => handlePageChange(data.pagination.page + 1)}
			>
				Next
			</Button>
		</div>
	{/if}
</div>
