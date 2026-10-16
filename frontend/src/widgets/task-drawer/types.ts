import type { Task } from '@/shared/api';

export interface TaskDrawerProps {
  task: Task | null;
  workspaceId?: string;
  onClose: () => void;
  onToggleTask?: (taskId: string) => void;
  onDeleteTask?: (taskId: string) => void;
}
