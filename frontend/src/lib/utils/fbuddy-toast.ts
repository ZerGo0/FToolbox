import { toast } from 'svelte-sonner';
import { browser } from '$app/environment';
import FBuddyToast from '$lib/components/FBuddyToast.svelte';

const STORAGE_KEY = 'fbuddy-toast-last-shown';
const COOLDOWN_HOURS = 24;

export function shouldShowFBuddyToast(): boolean {
  if (!browser) return false;

  const random = Math.random();
  if (random > 0.1) return false;

  const lastShown = localStorage.getItem(STORAGE_KEY);
  if (!lastShown) return true;

  const lastShownTime = parseInt(lastShown, 10);
  const hoursSinceLastShown = (Date.now() - lastShownTime) / (1000 * 60 * 60);

  return hoursSinceLastShown > COOLDOWN_HOURS;
}

export function showFBuddyToast() {
  if (!shouldShowFBuddyToast()) return;

  localStorage.setItem(STORAGE_KEY, Date.now().toString());

  toast.custom(FBuddyToast, {
    duration: 15000,
    position: 'bottom-right',
    dismissable: true
  });
}

export function resetFBuddyToastCooldown() {
  if (browser) {
    localStorage.removeItem(STORAGE_KEY);
  }
}
