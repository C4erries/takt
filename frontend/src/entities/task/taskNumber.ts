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
