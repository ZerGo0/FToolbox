<script lang="ts">
  import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import { RangeCalendar } from '$lib/components/ui/range-calendar/index.js';
  import * as Select from '$lib/components/ui/select/index.js';
  import { cn } from '$lib/utils.js';
  import { DateFormatter, type DateValue, getLocalTimeZone, today } from '@internationalized/date';
  import type { DateRange } from 'bits-ui';
  import CalendarIcon from 'lucide-svelte/icons/calendar';

  let {
    value = $bindable(),
    placeholder = 'Pick a date range',
    presets = [],
    onPresetSelect,
    onApply
  }: {
    value?: DateRange;
    placeholder?: string;
    presets?: Array<{ label: string; days: number }>;
    onPresetSelect?: (days: number) => void;
    onApply?: () => void;
  } = $props();

  const df = new DateFormatter('en-US', {
    dateStyle: 'medium'
  });

  let startValue: DateValue | undefined = $state(undefined);
  let open = $state(false);
</script>

<Popover.Root bind:open>
  <Popover.Trigger
    class={cn(
      buttonVariants({ variant: 'outline' }),
      cn('h-full justify-start text-left font-normal', !value && 'text-muted-foreground')
    )}
  >
    <CalendarIcon class="mr-2 size-4" />
    {#if value && value.start}
      {#if value.end}
        {df.format(value.start.toDate(getLocalTimeZone()))} - {df.format(
          value.end.toDate(getLocalTimeZone())
        )}
      {:else}
        {df.format(value.start.toDate(getLocalTimeZone()))}
      {/if}
    {:else if startValue}
      {df.format(startValue.toDate(getLocalTimeZone()))}
    {:else}
      {placeholder}
    {/if}
  </Popover.Trigger>
  <Popover.Content class="w-auto p-0" align="end">
    {#if presets.length > 0}
      <div class="p-3">
        <Select.Root
          type="single"
          onValueChange={(value) => {
            if (value && onPresetSelect) {
              onPresetSelect(parseInt(value));
            }
          }}
        >
          <Select.Trigger class="w-full">
            <span>Quick select</span>
          </Select.Trigger>
          <Select.Content>
            {#each presets as preset (preset.days)}
              <Select.Item value={preset.days.toString()}>{preset.label}</Select.Item>
            {/each}
          </Select.Content>
        </Select.Root>
      </div>
      <div class="border-t"></div>
    {/if}
    <RangeCalendar
      bind:value
      onStartValueChange={(v) => {
        startValue = v;
      }}
      numberOfMonths={2}
      placeholder={value?.start}
      maxValue={today(getLocalTimeZone())}
    />
    {#if onApply}
      <div class="border-t p-3">
        <Button
          class="w-full"
          size="sm"
          onclick={() => {
            if (onApply) {
              onApply();
              open = false;
            }
          }}
        >
          Apply
        </Button>
      </div>
    {/if}
  </Popover.Content>
</Popover.Root>
