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
  <AppSidebar />
  <main
    data-slot="sidebar-inset"
    class="bg-background relative flex w-full flex-1 flex-col overflow-hidden md:peer-data-[variant=inset]:m-2 md:peer-data-[variant=inset]:ml-0 md:peer-data-[variant=inset]:peer-data-[state=collapsed]:ml-2 md:peer-data-[variant=inset]:rounded-xl md:peer-data-[variant=inset]:shadow-sm"
  >
    <header
      class="h-14 group-has-data-[collapsible=icon]/sidebar-wrapper:h-14 flex shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear"
    >
      <div class="flex w-full items-center gap-1 px-4 lg:gap-2 lg:px-6">
        <SidebarTrigger />
        <Separator orientation="vertical" class="mx-2" />
        <DynamicBreadcrumbs />
      </div>
    </header>
    <div class="@container/main flex flex-1 flex-col gap-2 min-h-0 overflow-auto">
      <div class="p-6">
        {@render children()}
      </div>
    </div>
  </main>
</SidebarProvider>
