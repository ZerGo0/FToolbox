<script lang="ts">
  import { onMount } from 'svelte';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { CalendarIcon } from 'lucide-svelte';
  import Chart from 'chart.js/auto';
  import 'chartjs-adapter-date-fns';
  import type { DateRange } from 'bits-ui';
  import { getLocalTimeZone } from '@internationalized/date';

  interface Props {
    tagId: string;
    dateRange?: DateRange;
  }

  const { tagId, dateRange }: Props = $props();

  interface HistoryPoint {
    id: number;
    tagId: string;
    viewCount: number;
    createdAt: Date;
    updatedAt: Date;
  }

  let history = $state<HistoryPoint[]>([]);
  let loading = $state(true);
  let error = $state('');
  let startDate = $state(
    dateRange?.start?.toDate(getLocalTimeZone()) || new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)
  ); // 30 days ago
  let endDate = $state(dateRange?.end?.toDate(getLocalTimeZone()) || new Date());
  let chartCanvas = $state<HTMLCanvasElement>();
  let chartInstance: Chart | null = null;
  let showDatePicker = $state(false);

  // Use effect to update dates when dateRange changes
  $effect(() => {
    if (dateRange?.start && dateRange?.end) {
      const newStart = dateRange.start.toDate(getLocalTimeZone());
      const newEnd = dateRange.end.toDate(getLocalTimeZone());
      if (newStart.getTime() !== startDate.getTime() || newEnd.getTime() !== endDate.getTime()) {
        startDate = newStart;
        endDate = newEnd;
        // Fetch history after a small delay to avoid reactive loops
        setTimeout(() => {
          if (!loading) {
            fetchHistory();
          }
        }, 0);
      }
    }
  });

  async function fetchHistory() {
    loading = true;
    error = '';

    try {
      const params = new URLSearchParams({
        startDate: startDate.toISOString(),
        endDate: endDate.toISOString()
      });

      const response = await fetch(`http://localhost:3000/api/tags/${tagId}/history?${params}`);

      if (!response.ok) {
        throw new Error('Failed to fetch tag history');
      }

      const data = await response.json();
      history = data.history.map((point: Record<string, unknown>) => ({
        ...(point as unknown as HistoryPoint),
        createdAt: new Date(point.createdAt as string),
        updatedAt: new Date(point.updatedAt as string)
      }));

      updateChart();
    } catch (e) {
      error = e instanceof Error ? e.message : 'An error occurred';
    } finally {
      loading = false;
    }
  }

  function updateChart() {
    if (!chartCanvas) return;

    if (chartInstance) {
      chartInstance.destroy();
    }

    const chartData = history
      .map((point) => ({
        x: point.createdAt,
        y: point.viewCount
      }))
      .reverse();

    chartInstance = new Chart(chartCanvas, {
      type: 'line',
      data: {
        datasets: [
          {
            label: 'View Count',
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            data: chartData as any,
            borderColor: 'hsl(var(--primary))',
            backgroundColor: 'hsl(var(--primary) / 0.1)',
            tension: 0.1
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            display: false
          }
        },
        scales: {
          x: {
            type: 'time',
            time: {
              displayFormats: {
                day: 'MMM d'
              }
            }
          },
          y: {
            beginAtZero: false,
            ticks: {
              callback: function (value) {
                return new Intl.NumberFormat().format(value as number);
              }
            }
          }
        }
      }
    });
  }

  onMount(() => {
    // Only fetch history if component is mounted (i.e., row is expanded)
    // This prevents unnecessary API calls when the component is not visible
    fetchHistory();

    return () => {
      if (chartInstance) {
        chartInstance.destroy();
      }
    };
  });

  function formatDate(date: Date): string {
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function formatNumber(num: number): string {
    return new Intl.NumberFormat().format(num);
  }

  function updateDateRange() {
    showDatePicker = false;
    fetchHistory();
  }
</script>

<Card>
  <CardHeader>
    <CardTitle class="flex items-center justify-between">
      <span>Tag History</span>
      {#if !dateRange}
        <div class="relative">
          <Button variant="outline" size="sm" onclick={() => (showDatePicker = !showDatePicker)}>
            <CalendarIcon class="mr-2 h-4 w-4" />
            {formatDate(startDate)} - {formatDate(endDate)}
          </Button>

          {#if showDatePicker}
            <div
              class="bg-background absolute top-full right-0 z-50 mt-2 rounded-lg border p-4 shadow-lg"
            >
              <div class="space-y-4">
                <div>
                  <label for="start-date" class="text-sm font-medium">Start Date</label>
                  <input
                    id="start-date"
                    type="date"
                    value={startDate.toISOString().split('T')[0]}
                    onchange={(e) => (startDate = new Date(e.currentTarget.value))}
                    class="mt-1 block w-full rounded-md border px-3 py-2"
                  />
                </div>
                <div>
                  <label for="end-date" class="text-sm font-medium">End Date</label>
                  <input
                    id="end-date"
                    type="date"
                    value={endDate.toISOString().split('T')[0]}
                    onchange={(e) => (endDate = new Date(e.currentTarget.value))}
                    class="mt-1 block w-full rounded-md border px-3 py-2"
                  />
                </div>
                <div class="flex gap-2">
                  <Button size="sm" onclick={updateDateRange}>Apply</Button>
                  <Button size="sm" variant="outline" onclick={() => (showDatePicker = false)}
                    >Cancel</Button
                  >
                </div>
              </div>
            </div>
          {/if}
        </div>
      {:else}
        <span class="text-muted-foreground text-sm">
          {formatDate(startDate)} - {formatDate(endDate)}
        </span>
      {/if}
    </CardTitle>
  </CardHeader>
  <CardContent>
    {#if loading}
      <div class="flex justify-center py-8">
        <p class="text-muted-foreground">Loading history...</p>
      </div>
    {:else if error}
      <div class="flex justify-center py-8">
        <p class="text-destructive">{error}</p>
      </div>
    {:else if history.length === 0}
      <div class="flex justify-center py-8">
        <p class="text-muted-foreground">No history data available for this period</p>
      </div>
    {:else}
      <div class="space-y-4">
        <div class="h-64">
          <canvas bind:this={chartCanvas}></canvas>
        </div>

        <div class="space-y-2">
          <h4 class="text-sm font-medium">Data Points</h4>
          <div class="max-h-48 overflow-y-auto">
            <table class="w-full text-sm">
              <thead class="border-b">
                <tr>
                  <th class="py-2 text-left">Date</th>
                  <th class="py-2 text-right">View Count</th>
                  <th class="py-2 text-right">Change</th>
                </tr>
              </thead>
              <tbody>
                {#each history as point, i (point.id)}
                  <tr class="border-b">
                    <td class="py-2">{formatDate(point.createdAt)}</td>
                    <td class="py-2 text-right">{formatNumber(point.viewCount)}</td>
                    <td class="py-2 text-right">
                      {#if i < history.length - 1}
                        {@const change = point.viewCount - history[i + 1].viewCount}
                        <span class:text-green-600={change > 0} class:text-red-600={change < 0}>
                          {change > 0 ? '+' : ''}{formatNumber(change)}
                        </span>
                      {:else}
                        -
                      {/if}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    {/if}
  </CardContent>
</Card>
