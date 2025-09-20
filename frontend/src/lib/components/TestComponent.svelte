<script lang="ts">
  import { onMount } from 'svelte';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';

  onMount(() => {
    console.log('TestComponent mounted');
  });
</script>

<Card class="mb-6">
  <CardHeader>
    <CardTitle>Test Component</CardTitle>
  </CardHeader>
  <CardContent>
    <p class="text-muted-foreground text-sm">This is a test component.</p>
  </CardContent>
</Card>
