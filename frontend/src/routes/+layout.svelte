<script lang="ts">
  import { onNavigate } from '$app/navigation';
  import AppSidebar from '$lib/components/AppSidebar.svelte';
  import { SidebarProvider, SidebarTrigger } from '$lib/components/ui/sidebar';
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
<SidebarProvider>
  <div class="flex h-screen w-full">
    <AppSidebar />
    <main class="flex-1 overflow-auto">
      <div class="flex h-14 items-center border-b px-4">
        <SidebarTrigger />
      </div>
      <div class="p-6">
        {@render children()}
      </div>
    </main>
  </div>
</SidebarProvider>
