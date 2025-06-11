<script lang="ts">
  import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle
  } from '$lib/components/ui/card';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { TrendingUp, BarChart3, Clock, Tag } from 'lucide-svelte';
</script>

<div class="mx-auto max-w-4xl">
  <div class="mb-8">
    <h1 class="mb-4 text-4xl font-bold tracking-tight">Welcome to FanslyTagStats</h1>
    <p class="text-muted-foreground text-xl">
      Track and analyze Fansly tag popularity trends over time with automated data collection and
      insightful analytics.
    </p>
  </div>

  <div class="mb-8 grid gap-6 md:grid-cols-2">
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <TrendingUp class="h-5 w-5" />
          Real-time Tracking
        </CardTitle>
        <CardDescription>Monitor tag view counts and popularity changes</CardDescription>
      </CardHeader>
      <CardContent>
        <p class="text-muted-foreground text-sm">
          Our automated workers continuously collect data from Fansly, tracking view counts and
          calculating changes over time to help you understand tag performance trends.
        </p>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <BarChart3 class="h-5 w-5" />
          Historical Analytics
        </CardTitle>
        <CardDescription>Analyze trends with customizable date ranges</CardDescription>
      </CardHeader>
      <CardContent>
        <p class="text-muted-foreground text-sm">
          View historical data with flexible date range selection. Compare tag performance across
          different time periods to identify growth patterns and trending topics.
        </p>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <Clock class="h-5 w-5" />
          Automated Updates
        </CardTitle>
        <CardDescription>Set it and forget it with background workers</CardDescription>
      </CardHeader>
      <CardContent>
        <p class="text-muted-foreground text-sm">
          Background workers automatically update tag statistics daily, discover new trending tags,
          and maintain accurate historical records without manual intervention.
        </p>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <Tag class="h-5 w-5" />
          Tag Discovery
        </CardTitle>
        <CardDescription>Find new trending tags automatically</CardDescription>
      </CardHeader>
      <CardContent>
        <p class="text-muted-foreground text-sm">
          Our tag discovery system analyzes popular content to identify emerging tags, helping you
          stay ahead of trends and track new topics as they gain popularity.
        </p>
      </CardContent>
    </Card>
  </div>

  <Card class="border-primary">
    <CardHeader>
      <CardTitle>Get Started</CardTitle>
      <CardDescription>Begin tracking Fansly tags and analyzing their performance</CardDescription>
    </CardHeader>
    <CardContent class="space-y-4">
      <div class="flex items-start gap-4">
        <Badge class="mt-0.5">1</Badge>
        <div>
          <p class="font-medium">Browse existing tags</p>
          <p class="text-muted-foreground text-sm">
            Navigate to the Tags page to see currently tracked tags and their statistics.
          </p>
        </div>
      </div>

      <div class="flex items-start gap-4">
        <Badge class="mt-0.5">2</Badge>
        <div>
          <p class="font-medium">Add tags to track</p>
          <p class="text-muted-foreground text-sm">
            Use the "Add Tag" button to start tracking new tags. The system will fetch current data
            and begin monitoring changes.
          </p>
        </div>
      </div>

      <div class="flex items-start gap-4">
        <Badge class="mt-0.5">3</Badge>
        <div>
          <p class="font-medium">Analyze trends</p>
          <p class="text-muted-foreground text-sm">
            Click on any tag to view its historical data. Use date range filters to analyze
            performance over specific time periods.
          </p>
        </div>
      </div>

      <div class="pt-4">
        <Button href="/tags" size="lg">
          <Tag class="mr-2 h-4 w-4" />
          Go to Tags
        </Button>
      </div>
    </CardContent>
  </Card>
</div>
