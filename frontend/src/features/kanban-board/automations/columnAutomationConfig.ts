import type { Priority } from '@/shared/api';

export interface ColumnAutomationConfig {
  enterPriorityEnabled: boolean;
  enterPriority: Priority;
  completeMoveEnabled: boolean;
  completeTargetColumnId: string | null;
}

export const STORAGE_PREFIX = 'takt_col_auto_';

export function getStoredColumnAutomationConfig(colId: string, defId: string | null = null): ColumnAutomationConfig {
  try {
    const raw = localStorage.getItem(`${STORAGE_PREFIX}${colId}`);
    if (raw) return JSON.parse(raw);
  } catch {}
  return { enterPriorityEnabled: false, enterPriority: 'high', completeMoveEnabled: false, completeTargetColumnId: defId };
}

export function saveStoredColumnAutomationConfig(colId: string, cfg: ColumnAutomationConfig): void {
  try { localStorage.setItem(`${STORAGE_PREFIX}${colId}`, JSON.stringify(cfg)); } catch {}
}

export const PRIORITIES = [
  { value: 'none' as const, label: 'Обычный' }, { value: 'low' as const, label: 'Низкий' },
  { value: 'medium' as const, label: 'Средний' }, { value: 'high' as const, label: 'Высокий' },
  { value: 'pending' as const, label: 'В ожидании' },
];

export const PRIORITY_ACTIVE_STYLES: Record<string, string> = {
  pending: 'bg-[#4A6B82]/15 text-[#4A6B82] border-[#4A6B82]/40',
  urgent: 'bg-[#4A6B82]/15 text-[#4A6B82] border-[#4A6B82]/40',
  high: 'bg-[#D97746]/15 text-[#D97746] border-[#D97746]/40',
  medium: 'bg-[#B8943F]/15 text-[#B8943F] border-[#B8943F]/40',
  low: 'bg-[#566B44]/15 text-[#566B44] border-[#566B44]/40',
  none: 'bg-surface text-primary border-subtle',
};
