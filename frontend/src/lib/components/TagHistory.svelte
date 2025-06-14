<script lang="ts">
  import { onMount } from 'svelte';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import Chart from 'chart.js/auto';
  import 'chartjs-adapter-date-fns';

  interface HistoryPoint {
    id: number;
    tagId: string;
    viewCount: number;
    change: number;
    changePercent: number;
    createdAt: Date | string | number;
    updatedAt: Date | string | number;
  }

  interface Props {
    history?: HistoryPoint[];
  }

  const { history = [] }: Props = $props();
  let chartCanvas = $state<HTMLCanvasElement>();
  let chartInstance: Chart | null = null;
  let chartInitialized = false;

  // Convert dates in history to Date objects
  const historyWithDates = $derived(
    history.map((point) => ({
      ...point,
      createdAt:
        typeof point.createdAt === 'number'
          ? new Date(point.createdAt * 1000)
          : new Date(point.createdAt),
      updatedAt:
        typeof point.updatedAt === 'number'
          ? new Date(point.updatedAt * 1000)
          : new Date(point.updatedAt)
    }))
  );

  // Update chart when history changes
  $effect(() => {
    if (historyWithDates.length > 0) {
      setTimeout(() => updateChart(), 0);
    }
  });

  function updateChart() {
    if (!chartCanvas) {
      return;
    }

    if (chartInstance) {
      chartInstance.destroy();
      chartInstance = null;
    }

    // Always create chart data, even if empty
    const viewCountData =
      historyWithDates.length > 0
        ? historyWithDates
            .map((point) => ({
              x: point.createdAt,
              y: point.viewCount
            }))
            .reverse()
        : [];

    try {
      chartInstance = new Chart(chartCanvas, {
        type: 'line',
        data: {
          datasets: [
            {
              label: 'View Count',
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              data: viewCountData as any,
              borderColor: 'rgb(99, 102, 241)',
              backgroundColor: 'rgba(99, 102, 241, 0.1)',
              fill: true,
              tension: 0.4,
              pointRadius: 0,
              pointHoverRadius: 6,
              pointBackgroundColor: 'rgb(99, 102, 241)',
              pointBorderColor: '#fff',
              pointBorderWidth: 2,
              pointHoverBackgroundColor: 'rgb(99, 102, 241)',
              pointHoverBorderColor: '#fff',
              pointHoverBorderWidth: 2,
              yAxisID: 'y'
            }
          ]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          interaction: {
            mode: 'index',
            intersect: false
          },
          plugins: {
            legend: {
              display: true,
              position: 'top' as const,
              labels: {
                usePointStyle: true,
                padding: 20,
                font: {
                  size: 12
                }
              }
            },
            tooltip: {
              backgroundColor: 'rgba(0, 0, 0, 0.8)',
              titleColor: '#fff',
              bodyColor: '#fff',
              borderColor: 'rgb(99, 102, 241)',
              borderWidth: 1,
              padding: 12,
              displayColors: true,
              callbacks: {
                label: function (context) {
                  const label = context.dataset.label || '';
                  const value = new Intl.NumberFormat().format(context.parsed.y);
                  return label + ': ' + value;
                }
              }
            }
          },
          scales: {
            x: {
              type: 'time',
              time: {
                unit: 'day',
                displayFormats: {
                  day: 'MMM d'
                }
              },
              grid: {
                color: 'rgba(0, 0, 0, 0.05)',
                display: true
              },
              border: {
                display: false
              },
              ticks: {
                color: 'rgb(107, 114, 128)',
                font: {
                  size: 11
                }
              }
            },
            y: {
              type: 'linear',
              display: true,
              position: 'left',
              beginAtZero: false,
              grid: {
                color: 'rgba(0, 0, 0, 0.05)',
                display: true
              },
              border: {
                display: false
              },
              ticks: {
                color: 'rgb(107, 114, 128)',
                font: {
                  size: 11
                },
                callback: function (value) {
                  return new Intl.NumberFormat('en-US', {
                    notation: 'compact',
                    compactDisplay: 'short'
                  }).format(value as number);
                }
              },
              title: {
                display: true,
                text: 'View Count',
                color: 'rgb(107, 114, 128)',
                font: {
                  size: 12
                }
              }
            }
          }
        }
      });
      chartInitialized = true;
    } catch (error) {
      console.error('Failed to create chart:', error);
    }
  }

  // Watch for canvas element to be available
  $effect(() => {
    if (chartCanvas && !chartInitialized) {
      chartInitialized = true;
      updateChart();
    }
  });

  onMount(() => {
    return () => {
      if (chartInstance) {
        chartInstance.destroy();
        chartInstance = null;
        chartInitialized = false;
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
</script>

<Card>
  <CardHeader>
    <CardTitle>Tag History</CardTitle>
  </CardHeader>
  <CardContent>
    <div class="space-y-4">
      <div class="relative h-80">
        <canvas bind:this={chartCanvas} class="h-full w-full"></canvas>
        {#if historyWithDates.length === 0}
          <div class="pointer-events-none absolute inset-0 flex items-center justify-center">
            <p class="text-muted-foreground">No history data available for this period</p>
          </div>
        {/if}
      </div>

      {#if historyWithDates.length > 0}
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
                {#each historyWithDates as point (point.id)}
                  <tr class="border-b">
                    <td class="py-2">{formatDate(point.createdAt)}</td>
                    <td class="py-2 text-right">{formatNumber(point.viewCount)}</td>
                    <td class="py-2 text-right">
                      {#if point.change !== 0}
                        <span
                          class:text-green-600={point.change > 0}
                          class:text-red-600={point.change < 0}
                        >
                          {point.change > 0 ? '+' : ''}{formatNumber(point.change)} ({point.changePercent >=
                          0
                            ? '+'
                            : ''}{point.changePercent.toFixed(2)}%)
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
      {/if}
    </div>
  </CardContent>
</Card>
