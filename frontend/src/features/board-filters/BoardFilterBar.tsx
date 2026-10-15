import React from 'react';
import {
  Tag as TagIcon,
  X,
  ArrowUpDown,
  EyeOff,
  Calendar,
} from 'lucide-react';
import { useWorkspaceTagsQuery } from '@/entities/task';
import type { Priority, Tag } from '@/shared/api';
import type {
  BoardFilterBarProps,
  DeadlineFilter,
  SortOption,
} from './types';
import { initialFilterState } from './types';
import { isFilterActive } from './filterUtils';

const PRIORITIES: { value: Priority; label: string }[] = [
  { value: 'urgent', label: 'Срочно' },
  { value: 'high', label: 'Высокий' },
  { value: 'medium', label: 'Средний' },
  { value: 'low', label: 'Низкий' },
  { value: 'none', label: 'Обычный' },
];

const DEADLINES: { value: DeadlineFilter; label: string }[] = [
  { value: 'all', label: 'Все сроки' },
  { value: 'overdue', label: 'Просрочено' },
  { value: 'today', label: 'Сегодня' },
  { value: 'this_week', label: 'На неделе' },
  { value: 'no_date', label: 'Без срока' },
];

const SORT_OPTIONS: { value: SortOption; label: string }[] = [
  { value: 'manual', label: 'Вручную (DND)' },
  { value: 'due_date', label: 'По сроку' },
  { value: 'priority', label: 'По приоритету' },
  { value: 'created_at', label: 'По дате создания' },
];

export const BoardFilterBar: React.FC<BoardFilterBarProps> = ({
  filters,
  onFilterChange,
  workspaceId,
  className = '',
}) => {
  const { data: allTags = [] } = useWorkspaceTagsQuery(workspaceId);
  const active = isFilterActive(filters);

  const toggleTag = (tagId: string) => {
    const next = filters.tagIds.includes(tagId)
      ? filters.tagIds.filter((id) => id !== tagId)
      : [...filters.tagIds, tagId];
    onFilterChange({ ...filters, tagIds: next });
  };

  const togglePriority = (priority: Priority) => {
    const next = filters.priorities.includes(priority)
      ? filters.priorities.filter((p) => p !== priority)
      : [...filters.priorities, priority];
    onFilterChange({ ...filters, priorities: next });
  };

  const handleDeadlineChange = (deadline: DeadlineFilter) => {
    onFilterChange({ ...filters, deadline });
  };

  const handleSortChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onFilterChange({ ...filters, sortBy: e.target.value as SortOption });
  };

  const toggleHideCompleted = () => {
    onFilterChange({ ...filters, hideCompleted: !filters.hideCompleted });
  };

  const resetFilters = () => {
    onFilterChange(initialFilterState);
  };

  return (
    <div
      className={`bg-surface/50 border border-subtle/70 rounded-xl p-2.5 space-y-2 text-xs select-none shadow-washi-sm ${className}`}
    >
      {/* Top row: Tags & Reset */}
      <div className="flex items-center justify-between gap-2 flex-wrap">
        <div className="flex items-center gap-1.5 flex-wrap">
          <span className="text-[10px] font-mono uppercase tracking-wider text-muted flex items-center gap-1 mr-1">
            <TagIcon className="w-3 h-3" />
            <span>Теги (AND):</span>
          </span>

          {allTags.length === 0 ? (
            <span className="text-[11px] text-muted italic">нет тегов</span>
          ) : (
            allTags.map((tag: Tag) => {
              const isSelected = filters.tagIds.includes(tag.id);
              return (
                <button
                  key={tag.id}
                  type="button"
                  onClick={() => toggleTag(tag.id)}
                  className={`text-[11px] px-2 py-0.5 rounded-full border transition-all ${
                    isSelected
                      ? 'bg-accent-solid text-surface border-accent-solid font-medium shadow-washi-sm'
                      : 'border-subtle/80 bg-canvas/30 text-primary hover:bg-canvas/70'
                  }`}
                >
                  #{tag.name}
                </button>
              );
            })
          )}
        </div>

        {/* Reset button */}
        {active && (
          <button
            type="button"
            onClick={resetFilters}
            className="flex items-center gap-1 text-[11px] font-mono text-muted hover:text-primary transition-colors py-0.5 px-2 rounded-md hover:bg-canvas/50"
            title="Сбросить все фильтры"
          >
            <X className="w-3 h-3" />
            <span>Сбросить</span>
          </button>
        )}
      </div>

      {/* Bottom row: Priority, Deadline, Hide Completed, Sort */}
      <div className="flex items-center justify-between gap-3 flex-wrap pt-1.5 border-t border-subtle/40">
        <div className="flex items-center gap-3 flex-wrap">
          {/* Priority filter */}
          <div className="flex items-center gap-1">
            <span className="text-[10px] font-mono uppercase tracking-wider text-muted mr-0.5">
              Приоритет:
            </span>
            {PRIORITIES.map((p) => {
              const isSelected = filters.priorities.includes(p.value);
              return (
                <button
                  key={p.value}
                  type="button"
                  onClick={() => togglePriority(p.value)}
                  className={`text-[10px] px-1.5 py-0.5 rounded border transition-all ${
                    isSelected
                      ? 'bg-accent-surface border-accent-solid/60 text-primary font-medium'
                      : 'border-subtle/60 text-muted hover:text-primary hover:bg-canvas/40'
                  }`}
                >
                  {p.label}
                </button>
              );
            })}
          </div>

          {/* Deadline filter */}
          <div className="flex items-center gap-1">
            <span className="text-[10px] font-mono uppercase tracking-wider text-muted flex items-center gap-1 mr-0.5">
              <Calendar className="w-2.5 h-2.5" />
              Срок:
            </span>
            <select
              value={filters.deadline}
              onChange={(e) => handleDeadlineChange(e.target.value as DeadlineFilter)}
              className="text-[11px] font-mono bg-canvas/40 border border-subtle rounded px-1.5 py-0.5 text-primary focus:outline-none focus:border-accent-solid"
            >
              {DEADLINES.map((d) => (
                <option key={d.value} value={d.value}>
                  {d.label}
                </option>
              ))}
            </select>
          </div>

          {/* Hide completed toggle */}
          <button
            type="button"
            onClick={toggleHideCompleted}
            className={`flex items-center gap-1 px-2 py-0.5 rounded text-[11px] border transition-colors ${
              filters.hideCompleted
                ? 'bg-accent-surface border-accent-solid/60 text-primary font-medium'
                : 'border-subtle/60 text-muted hover:text-primary hover:bg-canvas/40'
            }`}
          >
            <EyeOff className="w-3 h-3" />
            <span>Без выполненных</span>
          </button>
        </div>

        {/* View Sorting selector */}
        <div className="flex items-center gap-1.5">
          <span className="text-[10px] font-mono uppercase tracking-wider text-muted flex items-center gap-1">
            <ArrowUpDown className="w-3 h-3" />
            <span>Сортировка:</span>
          </span>
          <select
            value={filters.sortBy}
            onChange={handleSortChange}
            className="text-[11px] font-mono bg-canvas/40 border border-subtle rounded px-2 py-0.5 text-primary focus:outline-none focus:border-accent-solid"
          >
            {SORT_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>
      </div>
    </div>
  );
};
