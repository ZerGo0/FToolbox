<script lang="ts">
  import { PUBLIC_API_URL } from '$env/static/public';
  import { onMount, onDestroy } from 'svelte';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Loader2, Play, AlertCircle } from 'lucide-svelte';

  interface WorkerStatus {
    id: number;
    name: string;
    lastRunAt: string | null;
    nextRunAt: string | null;
    status: 'idle' | 'running' | 'failed';
    lastError: string | null;
    runCount: number;
    successCount: number;
    failureCount: number;
    isEnabled: boolean;
    isRunning: boolean;
  }

  let workers: WorkerStatus[] = [];
  let loading = true;
  let error: string | null = null;
  let refreshInterval: number;

  async function fetchWorkerStatus() {
    try {
      const response = await fetch(`${PUBLIC_API_URL}/api/workers/status`);
      if (!response.ok) throw new Error('Failed to fetch worker status');

      const result = await response.json();
      workers = result.data;
      error = null;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to fetch worker status';
    } finally {
      loading = false;
    }
  }

  async function triggerWorker(workerName: string) {
    try {
      const response = await fetch(`${PUBLIC_API_URL}/api/workers/${workerName}/trigger`, {
        method: 'POST'
      });
      if (!response.ok) throw new Error('Failed to trigger worker');

      // Refresh status after trigger
      await fetchWorkerStatus();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to trigger worker';
    }
  }

  function formatDate(dateString: string | null): string {
    if (!dateString) return 'Never';
    const date = new Date(dateString);
    return date.toLocaleString();
  }

  function getStatusColor(status: string): 'default' | 'destructive' | 'secondary' {
    switch (status) {
      case 'running':
        return 'default';
      case 'failed':
        return 'destructive';
      default:
        return 'secondary';
    }
  }

  onMount(() => {
    fetchWorkerStatus();
    // Refresh every 10 seconds
    refreshInterval = setInterval(fetchWorkerStatus, 10000);
  });

  onDestroy(() => {
    clearInterval(refreshInterval);
  });
</script>

<Card class="mb-6">
  <CardHeader>
    <CardTitle>Worker Status</CardTitle>
  </CardHeader>
  <CardContent>
    {#if loading}
      <div class="flex items-center justify-center py-4">
        <Loader2 class="h-6 w-6 animate-spin" />
      </div>
    {:else if error}
      <div class="text-destructive flex items-center gap-2">
        <AlertCircle class="h-4 w-4" />
        <span>{error}</span>
      </div>
    {:else}
      <div class="space-y-4">
        {#each workers as worker (worker.id)}
          <div class="rounded-lg border p-4">
            <div class="mb-2 flex items-center justify-between">
              <div class="flex items-center gap-3">
                <h3 class="font-semibold capitalize">{worker.name.replace('-', ' ')}</h3>
                <Badge variant={getStatusColor(worker.status)}>
                  {#if worker.status === 'running'}
                    <Loader2 class="mr-1 h-3 w-3 animate-spin" />
                  {/if}
                  {worker.status}
                </Badge>
                {#if !worker.isEnabled}
                  <Badge variant="outline">Disabled</Badge>
                {/if}
              </div>
              <Button
                size="sm"
                variant="outline"
                disabled={worker.status === 'running' || !worker.isEnabled}
                onclick={() => triggerWorker(worker.name)}
              >
                <Play class="mr-1 h-3 w-3" />
                Run Now
              </Button>
            </div>

            <div class="grid grid-cols-2 gap-4 text-sm md:grid-cols-4">
              <div>
                <p class="text-muted-foreground">Last Run</p>
                <p class="font-mono">{formatDate(worker.lastRunAt)}</p>
              </div>
              <div>
                <p class="text-muted-foreground">Next Run</p>
                <p class="font-mono">{formatDate(worker.nextRunAt)}</p>
              </div>
              <div>
                <p class="text-muted-foreground">Success Rate</p>
                <p class="font-mono">
                  {worker.runCount > 0
                    ? `${Math.round((worker.successCount / worker.runCount) * 100)}%`
                    : 'N/A'}
                </p>
              </div>
              <div>
                <p class="text-muted-foreground">Total Runs</p>
                <p class="font-mono">{worker.runCount}</p>
              </div>
            </div>

            {#if worker.lastError}
              <div class="bg-destructive/10 mt-3 rounded p-2 text-sm">
                <p class="text-destructive font-semibold">Last Error:</p>
                <p class="text-destructive/80">{worker.lastError}</p>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </CardContent>
</Card>
