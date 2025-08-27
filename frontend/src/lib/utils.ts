import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Format a Date object to YYYY-MM-DD string using local timezone
 * This avoids the timezone issue with toISOString() which converts to UTC
 */
export function formatDateToLocal(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

/**
 * Format a date string or Date object to localized date and time
 */
export function formatDateTime(dateInput: string | Date, options?: Intl.DateTimeFormatOptions): string {
  const date = typeof dateInput === 'string' ? new Date(dateInput) : dateInput;
  
  if (isNaN(date.getTime())) {
    return '-';
  }

  const defaultOptions: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    ...options,
  };

  return new Intl.DateTimeFormat(navigator.language || 'en-US', defaultOptions).format(date);
}

/**
 * Format a date string or Date object to localized date only
 */
export function formatDate(dateInput: string | Date, options?: Intl.DateTimeFormatOptions): string {
  return formatDateTime(dateInput, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    ...options,
  });
}
