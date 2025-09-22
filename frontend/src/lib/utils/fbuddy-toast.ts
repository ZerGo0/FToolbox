import { browser } from '$app/environment';
import { toast } from 'svelte-sonner';

const STORAGE_KEY = 'fbuddy-toast-last-shown';
const COOLDOWN_HOURS = 24;

export function shouldShowFBuddyToast(): boolean {
  if (!browser) return false;

  // 15% chance to show
  //const random = Math.random();
  //if (random > 0.15) return false;

  // Check cooldown
  const lastShown = localStorage.getItem(STORAGE_KEY);
  if (!lastShown) return true;

  const lastShownTime = parseInt(lastShown, 10);
  const hoursSinceLastShown = (Date.now() - lastShownTime) / (1000 * 60 * 60);

  return hoursSinceLastShown > COOLDOWN_HOURS;
}

export function showFBuddyToast() {
  if (!shouldShowFBuddyToast()) return;

  localStorage.setItem(STORAGE_KEY, Date.now().toString());

  toast('FBuddy: Browser extension for Fansly creators', {
    duration: 12000,
    position: 'bottom-right',
    dismissable: true,
    closeButton: true,
    action: {
      label: 'Visit FBuddy.net',
      onClick: () => {
        // Keep referrer by omitting `noreferrer`, retain `noopener` for safety
        window.open('https://fbuddy.net/', '_blank', 'noopener');
      }
    }
  });
}

export function resetFBuddyToastCooldown() {
  if (browser) {
    localStorage.removeItem(STORAGE_KEY);
  }
}
