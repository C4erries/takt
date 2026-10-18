import React from 'react';
import { X } from 'lucide-react';
import type { Priority, Tag } from '@/shared/api';
import type { BoardFilterState, DeadlineFilter } from './types';

const PRIORITIES: Record<Priority, string> = {
  pending: 'В ожидании',
  urgent: 'В ожидании',
  high: 'Высокий',
  medium: 'Средний',
  low: 'Низкий',
  none: 'Обычный',
};

const DEADLINES: Record<DeadlineFilter, string> = {
  all: 'Все',
  overdue: 'Просрочено',
  today: 'Сегодня',
  this_week: 'На этой неделе',
  no_date: 'Без срока',
};

export interface FilterChipsProps {
  filters: BoardFilterState;
  allTags: Tag[];
  onRemovePriority: (p: Priority) => void;
  onRemoveDeadline: () => void;
  onRemoveTag: (tagId: string) => void;
  onToggleHideCompleted: () => void;
  onResetAll: () => void;
  active: boolean;
}

export const FilterChips: React.FC<FilterChipsProps> = ({
  filters,
  allTags,
  onRemovePriority,
  onRemoveDeadline,
  onRemoveTag,
  onToggleHideCompleted,
  onResetAll,
  active,
}) => {
  return (
    <>
      {/* Priority Chips */}
      {filters.priorities.map((p) => (
        <span
          key={p}
          className="inline-flex items-center gap-1 bg-surface border border-subtle/90 text-primary text-xs px-2.5 py-1 rounded-full shadow-washi-sm animate-in fade-in"
        >
          <span className="text-muted">Приоритет:</span>
          <span className="font-medium">{PRIORITIES[p] || p}</span>
          <button
            type="button"
            onClick={() => onRemovePriority(p)}
            className="text-muted hover:text-primary p-0.5 rounded-full hover:bg-canvas/50 transition-colors ml-0.5"
            title="Удалить фильтр приоритета"
          >
            <X className="w-3 h-3" />
          </button>
        </span>
      ))}

      {/* Deadline Chip */}
      {filters.deadline !== 'all' && (
        <span className="inline-flex items-center gap-1 bg-surface border border-subtle/90 text-primary text-xs px-2.5 py-1 rounded-full shadow-washi-sm animate-in fade-in">
          <span className="text-muted">Срок:</span>
          <span className="font-medium">
            {DEADLINES[filters.deadline] || filters.deadline}
          </span>
          <button
            type="button"
            onClick={onRemoveDeadline}
            className="text-muted hover:text-primary p-0.5 rounded-full hover:bg-canvas/50 transition-colors ml-0.5"
            title="Удалить фильтр срока"
          >
            <X className="w-3 h-3" />
          </button>
        </span>
      )}

      {/* Tags Chips */}
      {filters.tagIds.map((tagId) => {
        const tag = allTags.find((t) => t.id === tagId);
        return (
          <span
            key={tagId}
            className="inline-flex items-center gap-1 bg-surface border border-subtle/90 text-primary text-xs px-2.5 py-1 rounded-full shadow-washi-sm animate-in fade-in"
          >
            <span className="text-muted">Тег:</span>
            <span className="font-medium">{tag?.name || tagId}</span>
            <button
              type="button"
              onClick={() => onRemoveTag(tagId)}
              className="text-muted hover:text-primary p-0.5 rounded-full hover:bg-canvas/50 transition-colors ml-0.5"
              title="Удалить фильтр тега"
            >
              <X className="w-3 h-3" />
            </button>
          </span>
        );
      })}

      {/* Hide Completed Chip */}
      {filters.hideCompleted && (
        <span className="inline-flex items-center gap-1 bg-surface border border-subtle/90 text-primary text-xs px-2.5 py-1 rounded-full shadow-washi-sm animate-in fade-in">
          <span className="font-medium">Без выполненных</span>
          <button
            type="button"
            onClick={onToggleHideCompleted}
            className="text-muted hover:text-primary p-0.5 rounded-full hover:bg-canvas/50 transition-colors ml-0.5"
            title="Показать выполненные"
          >
            <X className="w-3 h-3" />
          </button>
        </span>
      )}

      {/* Reset All Button */}
      {active && (
        <button
          type="button"
          onClick={onResetAll}
          className="text-[11px] font-mono text-muted hover:text-primary px-2 py-0.5 rounded-md hover:bg-canvas/50 transition-colors"
          title="Сбросить все фильтры"
        >
          Сбросить
        </button>
      )}
    </>
  );
};
