<script lang="ts">
  import '../app.css';
  import { SidebarProvider, SidebarTrigger } from '$lib/components/ui/sidebar';
  import AppSidebar from '$lib/components/AppSidebar.svelte';

  let { children } = $props();
</script>

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
