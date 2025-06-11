<script lang="ts">
  import { Button } from '$lib/components/ui/button';
  import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle
  } from '$lib/components/ui/card';
  import { BarChart3, Clock, Tag, TrendingUp } from 'lucide-svelte';
</script>

<div>
  <div class="mb-8">
    <h1 class="mb-4 text-4xl font-bold tracking-tight">Welcome to FToolbox</h1>
    <p class="text-muted-foreground text-xl">
      A comprehensive collection of tools for Fansly creators and users. Analyze trends, track
      performance, and gain insights with our suite of specialized tools.
    </p>
  </div>

  <div class="mb-8">
    <h2 class="mb-6 text-2xl font-semibold">Available Tools</h2>

    <Card class="mb-6">
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <Tag class="h-5 w-5" />
          Tag Analytics
        </CardTitle>
        <CardDescription>Track and analyze Fansly tag popularity trends</CardDescription>
      </CardHeader>
      <CardContent>
        <p class="text-muted-foreground mb-4">
          A powerful analytics tool for monitoring Fansly tag performance. Track view counts,
          analyze historical trends, and discover emerging tags with our automated data collection
          system.
        </p>

        <div class="grid gap-4 md:grid-cols-2">
          <div class="flex items-start gap-2">
            <TrendingUp class="text-muted-foreground mt-0.5 h-4 w-4" />
            <div>
              <p class="text-sm font-medium">Real-time Tracking</p>
              <p class="text-muted-foreground text-xs">
                Monitor tag view counts and popularity changes
              </p>
            </div>
          </div>

          <div class="flex items-start gap-2">
            <BarChart3 class="text-muted-foreground mt-0.5 h-4 w-4" />
            <div>
              <p class="text-sm font-medium">Historical Analytics</p>
              <p class="text-muted-foreground text-xs">
                Analyze trends with customizable date ranges
              </p>
            </div>
          </div>

          <div class="flex items-start gap-2">
            <Clock class="text-muted-foreground mt-0.5 h-4 w-4" />
            <div>
              <p class="text-sm font-medium">Automated Updates</p>
              <p class="text-muted-foreground text-xs">Daily updates with background workers</p>
            </div>
          </div>

          <div class="flex items-start gap-2">
            <Tag class="text-muted-foreground mt-0.5 h-4 w-4" />
            <div>
              <p class="text-sm font-medium">Tag Discovery</p>
              <p class="text-muted-foreground text-xs">Find new trending tags automatically</p>
            </div>
          </div>
        </div>

        <div class="mt-4">
          <Button href="/tags" variant="default">
            <Tag class="mr-2 h-4 w-4" />
            Open Tag Analytics
          </Button>
        </div>
      </CardContent>
    </Card>

    <Card class="border-muted">
      <CardHeader>
        <CardTitle class="text-muted-foreground flex items-center gap-2">
          <Clock class="h-5 w-5" />
          More Tools Coming Soon
        </CardTitle>
        <CardDescription
          >We're working on additional tools to enhance your Fansly experience</CardDescription
        >
      </CardHeader>
      <CardContent>
        <p class="text-muted-foreground text-sm">
          Stay tuned for more powerful tools including content analytics, engagement metrics,
          scheduling tools, and much more. FToolbox is continuously evolving to provide the best
          toolkit for Fansly creators.
        </p>
      </CardContent>
    </Card>
  </div>
</div>
