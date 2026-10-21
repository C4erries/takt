import { useState, useEffect, useCallback } from 'react';

export const TASK_NUMBERS_STORAGE_KEY = 'takt_show_task_numbers';

/**
 * Formats a task ID into a short, consistent WEEEK-style number string (e.g. "#86").
 */
export function formatTaskNumber(taskId?: string | null): string {
  if (!taskId) return '#0';
  const digits = taskId.replace(/\D/g, '');
  if (digits.length >= 2) {
    return `#${digits.slice(-2)}`;
  }
  let hash = 0;
  for (let i = 0; i < taskId.length; i++) {
    hash = (hash * 33 + taskId.charCodeAt(i)) % 900;
  }
  return `#${Math.abs(hash) + 10}`;
}

export function getShowTaskNumbers(): boolean {
  if (typeof window === 'undefined') return false;
  try {
    return localStorage.getItem(TASK_NUMBERS_STORAGE_KEY) === 'true';
  } catch {
    return false;
  }
}

export function setShowTaskNumbers(show: boolean): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(TASK_NUMBERS_STORAGE_KEY, String(show));
    window.dispatchEvent(new Event('takt:show_task_numbers_changed'));
  } catch {
    // Ignore storage write issues
  }
}

export function useShowTaskNumbers(): [boolean, (val: boolean) => void] {
  const [show, setShow] = useState<boolean>(() => getShowTaskNumbers());

  useEffect(() => {
    const handleUpdate = () => setShow(getShowTaskNumbers());
    window.addEventListener('storage', handleUpdate);
    window.addEventListener('takt:show_task_numbers_changed', handleUpdate);
    return () => {
      window.removeEventListener('storage', handleUpdate);
      window.removeEventListener('takt:show_task_numbers_changed', handleUpdate);
    };
  }, []);

  const update = useCallback((val: boolean) => {
    setShow(val);
    setShowTaskNumbers(val);
  }, []);

  return [show, update];
}
