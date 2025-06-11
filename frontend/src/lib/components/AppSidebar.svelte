<script lang="ts">
  import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupContent,
    SidebarGroupLabel,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuItem,
    SidebarSeparator
  } from '$lib/components/ui/sidebar';
  import { Button } from '$lib/components/ui/button';
  import { Badge } from '$lib/components/ui/badge';
  import { Package2, Tag, Moon, Sun, Loader2 } from 'lucide-svelte';
  import { page } from '$app/stores';
  import { onMount } from 'svelte';

  let theme = $state('light');
  let workerStatus = $state<'idle' | 'running' | 'failed'>('idle');
  let checkingWorkerStatus = $state(false);

  // Check worker status
  async function checkWorkerStatus() {
    checkingWorkerStatus = true;
    try {
      const response = await fetch('http://localhost:3000/api/workers/status');
      if (response.ok) {
        const data = await response.json();
        workerStatus = data.status;
      }
    } catch (error) {
      console.error('Failed to check worker status:', error);
      workerStatus = 'failed';
    } finally {
      checkingWorkerStatus = false;
    }
  }

  // Toggle theme
  function toggleTheme() {
    theme = theme === 'light' ? 'dark' : 'light';
    document.documentElement.classList.toggle('dark');
    localStorage.setItem('theme', theme);
  }

  onMount(() => {
    // Load theme from localStorage
    const savedTheme = localStorage.getItem('theme') || 'light';
    theme = savedTheme;
    if (savedTheme === 'dark') {
      document.documentElement.classList.add('dark');
    }

    // Check worker status initially and periodically
    checkWorkerStatus();
    const interval = setInterval(checkWorkerStatus, 30000); // Check every 30 seconds

    return () => clearInterval(interval);
  });

  const menuItems = [
    { href: '/', label: 'Home', icon: Package2 },
    { href: '/tags', label: 'Tags', icon: Tag }
  ];
</script>

<Sidebar>
  <SidebarHeader>
    <SidebarMenu>
      <SidebarMenuItem>
        <a href="/" class="flex items-center gap-2 px-2 py-1.5 text-lg font-semibold">
          <Package2 class="h-6 w-6" />
          <span>FanslyTagStats</span>
        </a>
      </SidebarMenuItem>
    </SidebarMenu>
  </SidebarHeader>

  <SidebarContent>
    <SidebarGroup>
      <SidebarGroupLabel>Navigation</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {#each menuItems as item (item.href)}
            <SidebarMenuItem>
              <a
                href={item.href}
                class="hover:bg-accent hover:text-accent-foreground flex items-center gap-2 rounded-md px-2 py-1.5 text-sm {$page
                  .url.pathname === item.href
                  ? 'bg-accent text-accent-foreground'
                  : ''}"
              >
                <item.icon class="h-4 w-4" />
                <span>{item.label}</span>
              </a>
            </SidebarMenuItem>
          {/each}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  </SidebarContent>

  <SidebarFooter>
    <SidebarSeparator />

    <!-- Worker Status -->
    <div class="px-2 py-2">
      <div class="flex items-center justify-between rounded-md px-2 py-1.5 text-sm">
        <span class="text-muted-foreground">Worker Status</span>
        {#if checkingWorkerStatus}
          <Loader2 class="text-muted-foreground h-4 w-4 animate-spin" />
        {:else if workerStatus === 'running'}
          <Badge variant="default" class="h-5">Running</Badge>
        {:else if workerStatus === 'failed'}
          <Badge variant="destructive" class="h-5">Failed</Badge>
        {:else}
          <Badge variant="secondary" class="h-5">Idle</Badge>
        {/if}
      </div>
    </div>

    <SidebarSeparator />

    <!-- Theme Toggle -->
    <div class="p-2">
      <Button variant="ghost" size="sm" onclick={toggleTheme} class="w-full justify-start">
        {#if theme === 'light'}
          <Sun class="mr-2 h-4 w-4" />
          Light Mode
        {:else}
          <Moon class="mr-2 h-4 w-4" />
          Dark Mode
        {/if}
      </Button>
    </div>
  </SidebarFooter>
</Sidebar>
