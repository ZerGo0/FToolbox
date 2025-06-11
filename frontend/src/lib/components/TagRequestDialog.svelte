<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		Dialog,
		DialogContent,
		DialogDescription,
		DialogFooter,
		DialogHeader,
		DialogTitle,
		DialogTrigger
	} from '$lib/components/ui/dialog';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { Plus } from 'lucide-svelte';
	import { invalidateAll } from '$app/navigation';

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
			const response = await fetch('http://localhost:3000/api/tags/request', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ tag: tagInput.trim() })
			});

			const data = await response.json();

			if (!response.ok) {
				throw new Error(data.error || 'Failed to request tag');
			}

			success = data.message || 'Tag requested successfully';
			tagInput = '';

			// Refresh the page data
			await invalidateAll();

			// Close dialog after a short delay
			setTimeout(() => {
				open = false;
				success = '';
			}, 2000);
		} catch (e) {
			error = e instanceof Error ? e.message : 'An error occurred';
		} finally {
			loading = false;
		}
	}
</script>

<Dialog bind:open>
	<DialogTrigger>
		<Button>
			<Plus class="mr-2 h-4 w-4" />
			Request Tag
		</Button>
	</DialogTrigger>
	<DialogContent>
		<DialogHeader>
			<DialogTitle>Request New Tag</DialogTitle>
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
						placeholder="Enter tag name (without #)"
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
					{loading ? 'Requesting...' : 'Request Tag'}
				</Button>
			</DialogFooter>
		</form>
	</DialogContent>
</Dialog>
