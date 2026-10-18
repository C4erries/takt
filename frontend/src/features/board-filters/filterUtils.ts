import type { Task, Priority } from '@/shared/api';
import type { BoardFilterState } from './types';

const PRIORITY_WEIGHTS: Record<Priority, number> = {
  pending: 5,
  urgent: 5,
  high: 4,
  medium: 3,
  low: 2,
  none: 1,
};

export function isFilterActive(filters: BoardFilterState): boolean {
  return (
    filters.tagIds.length > 0 ||
    filters.priorities.length > 0 ||
    filters.deadline !== 'all' ||
    filters.hideCompleted ||
    filters.sortBy !== 'manual'
  );
}

function getTodayString(): string {
  const now = new Date();
  return now.toISOString().split('T')[0];
}

function getEndOfWeekString(): string {
  const now = new Date();
  const day = now.getDay();
  // Вычисляем дни до конца недели (воскресенье)
  const diff = now.getDate() + (day === 0 ? 0 : 7 - day);
  const endOfWeek = new Date(now.setDate(diff));
  return endOfWeek.toISOString().split('T')[0];
}

export function filterAndSortTasks(tasks: Task[], filters: BoardFilterState): Task[] {
  const todayStr = getTodayString();
  const endOfWeekStr = getEndOfWeekString();

  // 1. Фильтрация
  const filtered = tasks.filter((task) => {
    // Скрыть выполненные
    if (filters.hideCompleted && task.isCompleted) {
      return false;
    }

    // Фильтр тегов (Логика AND: задача обязана содержать ВСЕ выбранные теги)
    if (filters.tagIds.length > 0) {
      const taskTagIds = new Set(task.tags?.map((t) => t.id) ?? []);
      const hasAllTags = filters.tagIds.every((id) => taskTagIds.has(id));
      if (!hasAllTags) {
        return false;
      }
    }

    // Фильтр приоритета (любой из выбранных)
    if (filters.priorities.length > 0) {
      if (!filters.priorities.includes(task.priority)) {
        return false;
      }
    }

    // Фильтр дедлайна
    if (filters.deadline !== 'all') {
      if (filters.deadline === 'no_date') {
        if (task.dueDate) return false;
      } else if (filters.deadline === 'overdue') {
        if (!task.dueDate || task.dueDate >= todayStr || task.isCompleted) return false;
      } else if (filters.deadline === 'today') {
        if (task.dueDate !== todayStr) return false;
      } else if (filters.deadline === 'this_week') {
        if (!task.dueDate || task.dueDate < todayStr || task.dueDate > endOfWeekStr) return false;
      }
    }

    return true;
  });

  // 2. Сортировка отображения (View Sorting, не мутирует постоянные координаты position в базе)
  return filtered.sort((a, b) => {
    switch (filters.sortBy) {
      case 'due_date': {
        if (!a.dueDate && !b.dueDate) return a.position - b.position;
        if (!a.dueDate) return 1;
        if (!b.dueDate) return -1;
        const comp = a.dueDate.localeCompare(b.dueDate);
        return comp !== 0 ? comp : a.position - b.position;
      }
      case 'priority': {
        const weightA = PRIORITY_WEIGHTS[a.priority] ?? 0;
        const weightB = PRIORITY_WEIGHTS[b.priority] ?? 0;
        if (weightB !== weightA) return weightB - weightA;
        return a.position - b.position;
      }
      case 'created_at': {
        const timeA = new Date(a.createdAt).getTime();
        const timeB = new Date(b.createdAt).getTime();
        if (timeB !== timeA) return timeB - timeA;
        return a.position - b.position;
      }
      case 'manual':
      default:
        return a.position - b.position;
    }
  });
}
