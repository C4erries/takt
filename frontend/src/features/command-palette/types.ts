import type { Task, Project } from '@/shared/api';

export type CommandItemType = 'navigation' | 'action' | 'task' | 'project';

export interface CommandItem {
  id: string;
  type: CommandItemType;
  title: string;
  subtitle?: string;
  badge?: string;
  color?: string;
  isCompleted?: boolean;
  onSelect: () => void;
  taskData?: Task;
  projectData?: Project;
}

export interface CommandPaletteModalProps {
  isOpen: boolean;
  onClose: () => void;
  workspaceId?: string;
  onOpenQuickCapture?: () => void;
}
