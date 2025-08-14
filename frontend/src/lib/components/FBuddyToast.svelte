<script lang="ts">
  import { Button } from '$lib/components/ui/button';
  import { Card } from '$lib/components/ui/card';
  import { X, TrendingUp, Clock, DollarSign, BarChart3 } from 'lucide-svelte';

  let { onClose }: { onClose: () => void } = $props();

  const features = [
    {
      icon: TrendingUp,
      title: 'Unlimited FYP Analytics',
      description: 'Track your content performance with no limits'
    },
    {
      icon: DollarSign,
      title: 'DM Revenue Tracking',
      description: "See exactly how much you're earning from messages"
    },
    {
      icon: Clock,
      title: 'Fan Timezone Tracking',
      description: 'Know the perfect time to post for each fan'
    },
    {
      icon: BarChart3,
      title: 'Advanced Metrics',
      description: 'Detailed insights to 10x your growth'
    }
  ];

  const randomFeature = features[Math.floor(Math.random() * features.length)];
  const Icon = randomFeature.icon;

  const clickbaitMessages = [
    'Need help with your Fansly FYP? Try FBuddy',
    'Streamline your Fansly workflow with FBuddy Pro',
    'Track your Fansly performance better with FBuddy',
    'Want more insights into your Fansly content? FBuddy can help',
    'Simplify your Fansly analytics with FBuddy Pro'
  ];

  const randomMessage = clickbaitMessages[Math.floor(Math.random() * clickbaitMessages.length)];
</script>

<Card
  class="relative overflow-hidden border border-red-500/30 bg-gradient-to-br from-red-900/30 to-red-950/30 backdrop-blur-sm"
>
  <div
    class="pointer-events-none absolute inset-0 rounded-lg bg-gradient-to-br from-red-500/10 to-red-600/10"
  ></div>

  <Button
    variant="ghost"
    size="icon"
    class="absolute top-2 right-2 h-6 w-6 rounded-full hover:bg-white/10"
    onclick={onClose}
  >
    <X class="h-4 w-4" />
  </Button>

  <div class="p-4 pr-10">
    <div class="mb-3 flex items-start gap-3">
      <div class="rounded-lg bg-gradient-to-br from-[#ff6b6b] to-[#ff4757] p-2">
        <Icon class="h-5 w-5 text-white" />
      </div>
      <div class="flex-1">
        <div class="mb-1 flex items-center gap-2">
          <h3 class="text-sm font-bold text-white">FBuddy Pro</h3>
        </div>
        <p class="mb-2 text-xs font-medium text-yellow-400">{randomMessage}</p>
        <p class="mb-1 text-xs font-semibold text-white">{randomFeature.title}</p>
        <p class="mb-3 text-xs text-gray-300">{randomFeature.description}</p>

        <div class="flex flex-col gap-2">
          <div class="flex gap-2">
            <a
              href="https://chromewebstore.google.com/detail/fbuddy/odcgbfiaikgmamdjkiipobgolcpahmki"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex"
            >
              <Button
                size="sm"
                class="h-7 bg-gradient-to-r from-[#ff6b6b] to-[#ff4757] text-xs font-bold text-white hover:from-[#ff5252] hover:to-[#ff3838]"
              >
                Get FBuddy Pro
              </Button>
            </a>
            <Button
              size="sm"
              variant="ghost"
              class="h-7 text-xs text-gray-400 hover:text-white"
              onclick={onClose}
            >
              Not Now
            </Button>
          </div>
        </div>
      </div>
    </div>
  </div>

  <div
    class="absolute right-0 -bottom-1 left-0 h-1 bg-gradient-to-r from-[#ff6b6b] to-[#ff4757] opacity-75"
  ></div>
</Card>
