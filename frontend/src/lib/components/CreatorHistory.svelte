<script lang="ts">
  import { onMount } from 'svelte';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import Chart from 'chart.js/auto';
  import 'chartjs-adapter-date-fns';

  interface CreatorHistoryPoint {
    id: number;
    creatorId: string;
    mediaLikes: number;
    postLikes: number;
    followers: number;
    imageCount: number;
    videoCount: number;
    createdAt: number;
    updatedAt: number;
  }

  interface Props {
    history?: CreatorHistoryPoint[];
  }

  const { history = [] }: Props = $props();
  let chartCanvas = $state<HTMLCanvasElement>();
  let chartInstance: Chart | null = null;
  let chartInitialized = false;

  // Convert dates in history to Date objects
  const historyWithDates = $derived(
    history.map((point) => ({
      ...point,
      createdAt: new Date(point.createdAt * 1000),
      updatedAt: new Date(point.updatedAt * 1000)
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

    const datasets = [];

    // Prepare data for each metric
    const followersData = historyWithDates
      .map((point) => ({
        x: point.createdAt,
        y: point.followers
      }))
      .reverse();

    const mediaLikesData = historyWithDates
      .map((point) => ({
        x: point.createdAt,
        y: point.mediaLikes
      }))
      .reverse();

    const postLikesData = historyWithDates
      .map((point) => ({
        x: point.createdAt,
        y: point.postLikes
      }))
      .reverse();

    // Add all datasets
    datasets.push({
      label: 'Followers',
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      data: followersData as any,
      borderColor: 'rgb(99, 102, 241)',
      backgroundColor: 'rgba(99, 102, 241, 0.1)',
      fill: false,
      tension: 0.4,
      pointRadius: 0,
      pointHoverRadius: 6,
      pointBackgroundColor: 'rgb(99, 102, 241)',
      pointBorderColor: '#fff',
      pointBorderWidth: 2,
      pointHoverBackgroundColor: 'rgb(99, 102, 241)',
      pointHoverBorderColor: '#fff',
      pointHoverBorderWidth: 2,
      yAxisID: 'y-followers'
    });

    datasets.push({
      label: 'Media Likes',
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      data: mediaLikesData as any,
      borderColor: 'rgb(52, 211, 153)',
      backgroundColor: 'rgba(52, 211, 153, 0.1)',
      fill: false,
      tension: 0.4,
      pointRadius: 0,
      pointHoverRadius: 6,
      pointBackgroundColor: 'rgb(52, 211, 153)',
      pointBorderColor: '#fff',
      pointBorderWidth: 2,
      pointHoverBackgroundColor: 'rgb(52, 211, 153)',
      pointHoverBorderColor: '#fff',
      pointHoverBorderWidth: 2,
      yAxisID: 'y-likes'
    });

    datasets.push({
      label: 'Post Likes',
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      data: postLikesData as any,
      borderColor: 'rgb(251, 146, 60)',
      backgroundColor: 'rgba(251, 146, 60, 0.1)',
      fill: false,
      tension: 0.4,
      pointRadius: 0,
      pointHoverRadius: 6,
      pointBackgroundColor: 'rgb(251, 146, 60)',
      pointBorderColor: '#fff',
      pointBorderWidth: 2,
      pointHoverBackgroundColor: 'rgb(251, 146, 60)',
      pointHoverBorderColor: '#fff',
      pointHoverBorderWidth: 2,
      yAxisID: 'y-likes'
    });

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const scales: any = {
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
      }
    };

    // Configure scales for all metrics
    scales['y-followers'] = {
      type: 'linear',
      display: true,
      position: 'left',
      grid: {
        color: 'rgba(0, 0, 0, 0.05)',
        display: true
      },
      border: {
        display: false
      },
      ticks: {
        color: 'rgb(99, 102, 241)',
        font: {
          size: 11
        },
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        callback: function (value: any) {
          return new Intl.NumberFormat('en-US', {
            notation: 'compact',
            compactDisplay: 'short'
          }).format(value as number);
        }
      },
      title: {
        display: true,
        text: 'Followers',
        color: 'rgb(99, 102, 241)',
        font: {
          size: 12
        }
      }
    };

    scales['y-likes'] = {
      type: 'linear',
      display: true,
      position: 'right',
      grid: {
        drawOnChartArea: false
      },
      border: {
        display: false
      },
      ticks: {
        color: 'rgb(107, 114, 128)',
        font: {
          size: 11
        },
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        callback: function (value: any) {
          return new Intl.NumberFormat('en-US', {
            notation: 'compact',
            compactDisplay: 'short'
          }).format(value as number);
        }
      },
      title: {
        display: true,
        text: 'Likes',
        color: 'rgb(107, 114, 128)',
        font: {
          size: 12
        }
      }
    };

    try {
      chartInstance = new Chart(chartCanvas, {
        type: 'line',
        data: {
          datasets: datasets
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
                  const value = context.parsed.y;
                  return label + ': ' + new Intl.NumberFormat().format(value);
                }
              }
            }
          },
          scales: scales
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
    <CardTitle>Creator History</CardTitle>
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
                  <th class="py-2 text-right">Followers</th>
                  <th class="py-2 text-right">Media Likes</th>
                  <th class="py-2 text-right">Post Likes</th>
                </tr>
              </thead>
              <tbody>
                {#each historyWithDates as point (point.id)}
                  <tr class="border-b">
                    <td class="py-2">{formatDate(point.createdAt)}</td>
                    <td class="py-2 text-right">{formatNumber(point.followers)}</td>
                    <td class="py-2 text-right">{formatNumber(point.mediaLikes)}</td>
                    <td class="py-2 text-right">{formatNumber(point.postLikes)}</td>
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
