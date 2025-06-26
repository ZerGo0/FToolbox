import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, 'child'> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, 'children'> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };

// Number formatting utility
export function formatNumber(num: number): string {
  return new Intl.NumberFormat().format(num);
}

// Date formatting utility
export function formatDate(date: Date | string | number | null | undefined): string {
  if (!date) return 'Never';

  // Handle Unix timestamps (numbers)
  // If the number is less than 10 billion, it's likely in seconds, otherwise milliseconds
  const dateObj =
    typeof date === 'number'
      ? date < 1e10
        ? new Date(date * 1000)
        : new Date(date)
      : new Date(date);

  return dateObj.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  });
}

// Calculate change from history data
export function calculateChange(history: Array<{ value: number }> | undefined): {
  change: number;
  percentage: number;
} {
  if (!history || history.length === 0) {
    return { change: 0, percentage: 0 };
  }

  // History is in descending order (newest first)
  const newestValue = history[0].value;
  const oldestValue = history[history.length - 1].value;
  const totalChange = newestValue - oldestValue;
  const percentage = oldestValue > 0 ? (totalChange / oldestValue) * 100 : 0;

  return { change: totalChange, percentage };
}
