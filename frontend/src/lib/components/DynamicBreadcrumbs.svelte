<script lang="ts">
  import { page } from '$app/state';
  import * as Breadcrumb from '$lib/components/ui/breadcrumb';
  import { Home } from 'lucide-svelte';

  // Define breadcrumb labels for different routes
  const routeLabels: Record<string, string> = {
    '': 'Home',
    tags: 'Tags'
  };

  // Generate breadcrumb items from the current URL
  function generateBreadcrumbs() {
    const path = page.url.pathname;
    const segments = path.split('/').filter(Boolean);

    const items = [];
    let currentPath = '';

    // Always include home
    items.push({
      href: '/',
      label: 'Home',
      isHome: true
    });

    // Build breadcrumb items from URL segments
    segments.forEach((segment, index) => {
      currentPath += `/${segment}`;
      const isLast = index === segments.length - 1;

      // Get label from routeLabels or format the segment
      let label = routeLabels[segment] || decodeURIComponent(segment);

      items.push({
        href: isLast ? undefined : currentPath,
        label: label.charAt(0).toUpperCase() + label.slice(1),
        isHome: false
      });
    });

    return items;
  }

  let breadcrumbs = $derived(generateBreadcrumbs());
</script>

{#if breadcrumbs.length > 1}
  <Breadcrumb.Root>
    <Breadcrumb.List>
      {#each breadcrumbs as item, i (item.href || item.label)}
        <Breadcrumb.Item>
          {#if item.href}
            <Breadcrumb.Link href={item.href} class="flex items-center gap-1">
              {#if item.isHome}
                <Home class="h-4 w-4" />
              {:else}
                {item.label}
              {/if}
            </Breadcrumb.Link>
          {:else}
            <Breadcrumb.Page>{item.label}</Breadcrumb.Page>
          {/if}
        </Breadcrumb.Item>
        {#if i < breadcrumbs.length - 1}
          <Breadcrumb.Separator />
        {/if}
      {/each}
    </Breadcrumb.List>
  </Breadcrumb.Root>
{/if}
