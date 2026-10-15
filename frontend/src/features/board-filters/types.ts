import type { Priority } from '@/shared/api';

export type DeadlineFilter = 'all' | 'overdue' | 'today' | 'this_week' | 'no_date';

export type SortOption = 'manual' | 'due_date' | 'priority' | 'created_at';

export interface BoardFilterState {
  tagIds: string[];
  priorities: Priority[];
  deadline: DeadlineFilter;
  hideCompleted: boolean;
  sortBy: SortOption;
}

export const initialFilterState: BoardFilterState = {
  tagIds: [],
  priorities: [],
  deadline: 'all',
  hideCompleted: false,
  sortBy: 'manual',
};

export interface BoardFilterBarProps {
  filters: BoardFilterState;
  onFilterChange: (filters: BoardFilterState) => void;
  workspaceId?: string;
  className?: string;
}
