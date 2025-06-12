<script lang="ts">
  import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle
  } from '$lib/components/ui/card';
  import { Clock, Tag } from 'lucide-svelte';
</script>

<div>
  <div class="mb-8">
    <h1 class="mb-4 text-4xl font-bold tracking-tight">Welcome to FToolbox</h1>
    <p class="text-muted-foreground text-xl">
      A comprehensive collection of free tools for Fansly creators and users. Analyze trends, track
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
    </Card>

    <Card class="border-muted">
      <CardHeader>
        <CardTitle class="text-muted-foreground flex items-center gap-2">
          <Clock class="h-5 w-5" />
          More Tools Coming Soon
        </CardTitle>
        <CardDescription
          >Feel free to request new tools or features you'd like to see!</CardDescription
        >
      </CardHeader>
      <CardContent>
        <p class="text-muted-foreground text-sm">
          Join the
          <a
            href="https://discord.gg/SBWkgGcmfr"
            target="_blank"
            class="text-blue-500 hover:underline">Discord</a
          >
          server to stay updated on new features and get support!
        </p>
      </CardContent>
    </Card>
  </div>
</div>
