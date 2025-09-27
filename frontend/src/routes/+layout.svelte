<script lang="ts">
  import { onNavigate } from '$app/navigation';
  import AppSidebar from '$lib/components/AppSidebar.svelte';
  import DynamicBreadcrumbs from '$lib/components/DynamicBreadcrumbs.svelte';
  import { Separator } from '$lib/components/ui/separator';
  import { SidebarProvider, SidebarTrigger } from '$lib/components/ui/sidebar';
  import { Toaster } from '$lib/components/ui/sonner';
  import { ModeWatcher } from 'mode-watcher';
  import '../app.css';

  let { children } = $props();

  onNavigate((navigation) => {
    if (!document.startViewTransition) return;

    return new Promise((resolve) => {
      document.startViewTransition(async () => {
        resolve();
        await navigation.complete;
      });
    });
  });
</script>

<ModeWatcher />
<Toaster />
<SidebarProvider>
  <div class="flex min-h-0 w-full flex-1">
    <AppSidebar />
    <main class="flex min-h-0 flex-1 overflow-y-auto">
      <div class="flex h-14 items-center gap-4 border-b px-4">
        <SidebarTrigger />
        <Separator orientation="vertical" />
        <DynamicBreadcrumbs />
      </div>
      <div class="p-6">
        {@render children()}
      </div>
    </main>
  </div>
</SidebarProvider>
