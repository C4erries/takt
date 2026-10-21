import type { Task, Priority } from '@/shared/api';

export interface TaskDrawerProps {
  task: Task | null;
  workspaceId?: string;
  onClose: () => void;
  onToggleTask?: (taskId: string) => void;
  onDeleteTask?: (taskId: string) => void;
  onSelectTask?: (taskId: string) => void;
}

export const PRIORITIES: { value: Priority; label: string }[] = [
  { value: 'none', label: 'Обычный' },
  { value: 'low', label: 'Низкий' },
  { value: 'medium', label: 'Средний' },
  { value: 'high', label: 'Высокий' },
  { value: 'pending', label: 'В ожидании' },
];

export const PRIORITY_ACTIVE_STYLES: Record<string, string> = {
  pending: 'bg-accent-denim/15 text-accent-denim border-accent-denim/40',
  urgent: 'bg-accent-denim/15 text-accent-denim border-accent-denim/40',
  high: 'bg-accent-persimmon/15 text-accent-persimmon border-accent-persimmon/40',
  medium: 'bg-accent-straw/15 text-accent-straw border-accent-straw/40',
  low: 'bg-accent-solid/15 text-accent-solid border-accent-solid/40',
  none: 'bg-surface text-primary border-subtle',
};
