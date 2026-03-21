<script lang="ts">
  import { onMount } from 'svelte';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import Chart from 'chart.js/auto';
  import type { ChartConfiguration, ChartDataset } from 'chart.js';
  import 'chartjs-adapter-date-fns';

  interface HistoryPoint {
    id: number;
    tagId: string;
    viewCount: number;
    change: number;
    changePercent: number;
    postCount: number;
    ratio: number;
    postCountChange: number;
    createdAt: Date | string | number;
    updatedAt: Date | string | number;
  }

  interface HistoryPointWithDates extends Omit<HistoryPoint, 'createdAt' | 'updatedAt'> {
    createdAt: Date;
    updatedAt: Date;
  }

  interface Props {
    history?: HistoryPoint[];
  }

  type ChartPoint = { x: Date; y: number };

  const { history = [] }: Props = $props();
  let chartCanvas = $state<HTMLCanvasElement>();
  let chartInstance: Chart<'line', ChartPoint[]> | null = null;
  let chartInitialized = false;

  const historyWithDates = $derived<HistoryPointWithDates[]>(
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

  $effect(() => {
    if (chartCanvas) {
      setTimeout(() => updateChart(), 0);
    }
  });

  function buildSeries(
    metric: keyof Pick<
      HistoryPointWithDates,
      'viewCount' | 'change' | 'postCount' | 'ratio' | 'postCountChange'
    >
  ): ChartPoint[] {
    return historyWithDates
      .map((point) => ({
        x: point.createdAt,
        y: point[metric]
      }))
      .reverse();
  }

  function updateChart() {
    if (!chartCanvas) {
      return;
    }

    if (chartInstance) {
      chartInstance.destroy();
      chartInstance = null;
    }

    const datasets: ChartDataset<'line', ChartPoint[]>[] = [
      {
        label: 'View Count',
        data: buildSeries('viewCount'),
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
      },
      {
        label: 'Views Change',
        data: buildSeries('change'),
        borderColor: 'rgb(52, 211, 153)',
        backgroundColor: 'rgba(52, 211, 153, 0.15)',
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
        borderDash: [4, 4],
        hidden: true,
        yAxisID: 'y1'
      },
      {
        label: 'Post Count',
        data: buildSeries('postCount'),
        borderColor: 'rgb(251, 146, 60)',
        backgroundColor: 'rgba(251, 146, 60, 0.1)',
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHoverRadius: 6,
        pointBackgroundColor: 'rgb(251, 146, 60)',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        pointHoverBackgroundColor: 'rgb(251, 146, 60)',
        pointHoverBorderColor: '#fff',
        pointHoverBorderWidth: 2,
        yAxisID: 'y2'
      },
      {
        label: 'Ratio',
        data: buildSeries('ratio'),
        borderColor: 'rgb(236, 72, 153)',
        backgroundColor: 'rgba(236, 72, 153, 0.12)',
        fill: false,
        tension: 0.4,
        pointRadius: 0,
        pointHoverRadius: 6,
        pointBackgroundColor: 'rgb(236, 72, 153)',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        pointHoverBackgroundColor: 'rgb(236, 72, 153)',
        pointHoverBorderColor: '#fff',
        pointHoverBorderWidth: 2,
        yAxisID: 'y3'
      },
      {
        label: 'Post Change',
        data: buildSeries('postCountChange'),
        borderColor: 'rgb(14, 165, 233)',
        backgroundColor: 'rgba(14, 165, 233, 0.12)',
        fill: false,
        tension: 0.4,
        pointRadius: 0,
        pointHoverRadius: 6,
        pointBackgroundColor: 'rgb(14, 165, 233)',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        pointHoverBackgroundColor: 'rgb(14, 165, 233)',
        pointHoverBorderColor: '#fff',
        pointHoverBorderWidth: 2,
        borderDash: [8, 4],
        hidden: true,
        yAxisID: 'y1'
      }
    ];

    const chartConfig: ChartConfiguration<'line', ChartPoint[]> = {
      type: 'line',
      data: {
        datasets
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
            position: 'top',
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
              label: (context) => {
                const label = context.dataset.label || '';
                const value = context.parsed.y;
                if (value === null) {
                  return `${label}: N/A`;
                }
                if (label === 'Ratio') {
                  return `${label}: ${value.toFixed(2)}`;
                }
                if (label === 'Views Change' || label === 'Post Change') {
                  const formatted = new Intl.NumberFormat('en-US', {
                    signDisplay: 'exceptZero'
                  }).format(value);
                  return `${label}: ${formatted}`;
                }
                return `${label}: ${new Intl.NumberFormat().format(value)}`;
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
              callback: (value) =>
                new Intl.NumberFormat('en-US', {
                  notation: 'compact',
                  compactDisplay: 'short'
                }).format(Number(value))
            },
            title: {
              display: true,
              text: 'View Count',
              color: 'rgb(107, 114, 128)',
              font: {
                size: 12
              }
            }
          },
          y1: {
            type: 'linear',
            display: true,
            position: 'right',
            beginAtZero: false,
            grid: {
              drawOnChartArea: false
            },
            border: {
              display: false
            },
            ticks: {
              color: 'rgb(52, 211, 153)',
              font: {
                size: 11
              },
              callback: (value) =>
                new Intl.NumberFormat('en-US', {
                  signDisplay: 'exceptZero'
                }).format(Number(value))
            },
            title: {
              display: true,
              text: 'Changes',
              color: 'rgb(52, 211, 153)',
              font: {
                size: 12
              }
            }
          },
          y2: {
            type: 'linear',
            display: true,
            position: 'right',
            beginAtZero: false,
            grid: {
              drawOnChartArea: false
            },
            border: {
              display: false
            },
            ticks: {
              color: 'rgb(251, 146, 60)',
              font: {
                size: 11
              },
              callback: (value) =>
                new Intl.NumberFormat('en-US', {
                  notation: 'compact',
                  compactDisplay: 'short'
                }).format(Number(value))
            },
            title: {
              display: true,
              text: 'Post Count',
              color: 'rgb(251, 146, 60)',
              font: {
                size: 12
              }
            }
          },
          y3: {
            type: 'linear',
            display: true,
            position: 'right',
            offset: true,
            beginAtZero: false,
            grid: {
              drawOnChartArea: false
            },
            border: {
              display: false
            },
            ticks: {
              color: 'rgb(236, 72, 153)',
              font: {
                size: 11
              },
              callback: (value) => Number(value).toFixed(2)
            },
            title: {
              display: true,
              text: 'Ratio',
              color: 'rgb(236, 72, 153)',
              font: {
                size: 12
              }
            }
          }
        }
      }
    };

    try {
      chartInstance = new Chart(chartCanvas, chartConfig);
      chartInitialized = true;
    } catch (error) {
      console.error('Failed to create chart:', error);
    }
  }

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

  function formatRatio(ratio: number, postCount: number): string {
    if (postCount <= 0) {
      return '-';
    }

    return ratio.toFixed(2);
  }
</script>

<Card>
  <CardHeader>
    <CardTitle>Tag History</CardTitle>
  </CardHeader>
  <CardContent>
    <div class="space-y-4">
      <div class="relative h-80 w-full">
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
                  <th class="py-2 text-right">Views Change</th>
                  <th class="py-2 text-right">Post Count</th>
                  <th class="py-2 text-right">Ratio</th>
                  <th class="py-2 text-right">Post Change</th>
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
                    <td class="py-2 text-right">{formatNumber(point.postCount)}</td>
                    <td class="py-2 text-right">{formatRatio(point.ratio, point.postCount)}</td>
                    <td class="py-2 text-right">
                      {#if point.postCountChange !== 0}
                        <span
                          class:text-green-600={point.postCountChange > 0}
                          class:text-red-600={point.postCountChange < 0}
                        >
                          {point.postCountChange > 0 ? '+' : ''}{formatNumber(
                            point.postCountChange
                          )}
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
