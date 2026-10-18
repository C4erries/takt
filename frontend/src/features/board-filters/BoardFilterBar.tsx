import React, { useState, useRef, useEffect } from 'react';
import {
  Plus,
  ArrowUpDown,
  ChevronRight,
  ChevronLeft,
  Calendar,
  AlertCircle,
  Tag as TagIcon,
  CheckCircle2,
  Check,
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
import { FilterChips } from './FilterChips';

const PRIORITIES: { value: Priority; label: string }[] = [
  { value: 'urgent', label: 'Срочно' },
  { value: 'high', label: 'Высокий' },
  { value: 'medium', label: 'Средний' },
  { value: 'low', label: 'Низкий' },
  { value: 'none', label: 'Обычный' },
];

const DEADLINES: { value: DeadlineFilter; label: string }[] = [
  { value: 'overdue', label: 'Просрочено' },
  { value: 'today', label: 'Сегодня' },
  { value: 'this_week', label: 'На этой неделе' },
  { value: 'no_date', label: 'Без срока' },
];

const SORT_OPTIONS: { value: SortOption; label: string }[] = [
  { value: 'manual', label: 'По умолчанию (DND)' },
  { value: 'due_date', label: 'По сроку' },
  { value: 'priority', label: 'По приоритету' },
  { value: 'created_at', label: 'По дате создания' },
];

type CategoryMenu = 'root' | 'priority' | 'deadline' | 'tags';

export const BoardFilterBar: React.FC<BoardFilterBarProps> = ({
  filters,
  onFilterChange,
  workspaceId,
  className = '',
}) => {
  const { data: allTags = [] } = useWorkspaceTagsQuery(workspaceId);
  const [isOpen, setIsOpen] = useState(false);
  const [currentMenu, setCurrentMenu] = useState<CategoryMenu>('root');
  const dropdownRef = useRef<HTMLDivElement>(null);

  const active = isFilterActive(filters);

  useEffect(() => {
    if (!isOpen) return;
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setIsOpen(false);
        setCurrentMenu('root');
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  const togglePriority = (p: Priority) => {
    const exists = filters.priorities.includes(p);
    const next = exists
      ? filters.priorities.filter((item) => item !== p)
      : [...filters.priorities, p];
    onFilterChange({ ...filters, priorities: next });
    setIsOpen(false);
    setCurrentMenu('root');
  };

  const removePriority = (p: Priority) => {
    onFilterChange({
      ...filters,
      priorities: filters.priorities.filter((item) => item !== p),
    });
  };

  const setDeadline = (d: DeadlineFilter) => {
    onFilterChange({
      ...filters,
      deadline: filters.deadline === d ? 'all' : d,
    });
    setIsOpen(false);
    setCurrentMenu('root');
  };

  const toggleTag = (tagId: string) => {
    const exists = filters.tagIds.includes(tagId);
    const next = exists
      ? filters.tagIds.filter((id) => id !== tagId)
      : [...filters.tagIds, tagId];
    onFilterChange({ ...filters, tagIds: next });
    setIsOpen(false);
    setCurrentMenu('root');
  };

  const removeTag = (tagId: string) => {
    onFilterChange({
      ...filters,
      tagIds: filters.tagIds.filter((id) => id !== tagId),
    });
  };

  const toggleHideCompleted = () => {
    onFilterChange({
      ...filters,
      hideCompleted: !filters.hideCompleted,
    });
    setIsOpen(false);
    setCurrentMenu('root');
  };

  return (
    <div className={`flex items-center justify-between gap-2 min-h-9 flex-wrap text-xs select-none py-1 ${className}`}>
      <div className="flex items-center gap-1.5 flex-wrap">
        <div className="relative" ref={dropdownRef}>
          <button
            type="button"
            onClick={() => {
              setIsOpen((prev) => !prev);
              setCurrentMenu('root');
            }}
            className="bg-surface/60 hover:bg-surface border border-subtle/80 text-muted hover:text-primary rounded-lg px-2.5 py-1 text-xs flex items-center gap-1.5 transition-colors shadow-washi-sm"
            title="Добавить фильтр"
          >
            <Plus className="w-3.5 h-3.5" />
            <span>Добавить фильтр</span>
          </button>

          {isOpen && (
            <div className="absolute left-0 top-full mt-1.5 w-56 z-30 bg-surface border border-subtle rounded-xl shadow-washi p-1.5 space-y-0.5 text-xs animate-in fade-in zoom-in-95 duration-100">
              {currentMenu === 'root' && (
                <>
                  <button
                    type="button"
                    onClick={() => setCurrentMenu('priority')}
                    className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors"
                  >
                    <div className="flex items-center gap-2">
                      <AlertCircle className="w-3.5 h-3.5 text-muted" />
                      <span>Приоритет</span>
                    </div>
                    <ChevronRight className="w-3.5 h-3.5 text-muted" />
                  </button>

                  <button
                    type="button"
                    onClick={() => setCurrentMenu('deadline')}
                    className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors"
                  >
                    <div className="flex items-center gap-2">
                      <Calendar className="w-3.5 h-3.5 text-muted" />
                      <span>Срок</span>
                    </div>
                    <ChevronRight className="w-3.5 h-3.5 text-muted" />
                  </button>

                  <button
                    type="button"
                    onClick={() => setCurrentMenu('tags')}
                    className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors"
                  >
                    <div className="flex items-center gap-2">
                      <TagIcon className="w-3.5 h-3.5 text-muted" />
                      <span>Теги</span>
                    </div>
                    <ChevronRight className="w-3.5 h-3.5 text-muted" />
                  </button>

                  <div className="border-t border-subtle/40 my-1" />

                  <button
                    type="button"
                    onClick={toggleHideCompleted}
                    className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors"
                  >
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className="w-3.5 h-3.5 text-muted" />
                      <span>Скрыть выполненные</span>
                    </div>
                    {filters.hideCompleted && <Check className="w-3.5 h-3.5 text-accent-solid" />}
                  </button>
                </>
              )}

              {currentMenu === 'priority' && (
                <div className="space-y-0.5">
                  <button
                    type="button"
                    onClick={() => setCurrentMenu('root')}
                    className="w-full flex items-center gap-1.5 px-2 py-1 text-muted hover:text-primary text-[11px] font-mono mb-1"
                  >
                    <ChevronLeft className="w-3 h-3" />
                    <span>Назад</span>
                  </button>
                  {PRIORITIES.map((p) => (
                    <button
                      key={p.value}
                      type="button"
                      onClick={() => togglePriority(p.value)}
                      className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors"
                    >
                      <span>{p.label}</span>
                      {filters.priorities.includes(p.value) && (
                        <Check className="w-3.5 h-3.5 text-accent-solid" />
                      )}
                    </button>
                  ))}
                </div>
              )}

              {currentMenu === 'deadline' && (
                <div className="space-y-0.5">
                  <button
                    type="button"
                    onClick={() => setCurrentMenu('root')}
                    className="w-full flex items-center gap-1.5 px-2 py-1 text-muted hover:text-primary text-[11px] font-mono mb-1"
                  >
                    <ChevronLeft className="w-3 h-3" />
                    <span>Назад</span>
                  </button>
                  {DEADLINES.map((d) => (
                    <button
                      key={d.value}
                      type="button"
                      onClick={() => setDeadline(d.value)}
                      className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors"
                    >
                      <span>{d.label}</span>
                      {filters.deadline === d.value && (
                        <Check className="w-3.5 h-3.5 text-accent-solid" />
                      )}
                    </button>
                  ))}
                </div>
              )}

              {currentMenu === 'tags' && (
                <div className="space-y-0.5">
                  <button
                    type="button"
                    onClick={() => setCurrentMenu('root')}
                    className="w-full flex items-center gap-1.5 px-2 py-1 text-muted hover:text-primary text-[11px] font-mono mb-1"
                  >
                    <ChevronLeft className="w-3 h-3" />
                    <span>Назад</span>
                  </button>
                  {allTags.length === 0 ? (
                    <div className="px-2.5 py-2 text-muted italic text-[11px]">Нет тегов в воркспейсе</div>
                  ) : (
                    allTags.map((tag: Tag) => (
                      <button
                        key={tag.id}
                        type="button"
                        onClick={() => toggleTag(tag.id)}
                        className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors"
                      >
                        <span className="truncate">#{tag.name}</span>
                        {filters.tagIds.includes(tag.id) && (
                          <Check className="w-3.5 h-3.5 text-accent-solid" />
                        )}
                      </button>
                    ))
                  )}
                </div>
              )}
            </div>
          )}
        </div>

        <FilterChips
          filters={filters}
          allTags={allTags}
          onRemovePriority={removePriority}
          onRemoveDeadline={() => onFilterChange({ ...filters, deadline: 'all' })}
          onRemoveTag={removeTag}
          onToggleHideCompleted={() => onFilterChange({ ...filters, hideCompleted: !filters.hideCompleted })}
          onResetAll={() => onFilterChange(initialFilterState)}
          active={active}
        />
      </div>

      <div className="flex items-center gap-1.5 ml-auto">
        <ArrowUpDown className="w-3 h-3 text-muted" />
        <select
          value={filters.sortBy}
          onChange={(e) => onFilterChange({ ...filters, sortBy: e.target.value as SortOption })}
          className="text-xs bg-surface/60 hover:bg-surface border border-subtle/80 text-primary rounded-lg px-2 py-1 focus:outline-none focus:border-accent-solid transition-colors shadow-washi-sm cursor-pointer"
          title="Сортировка задач"
        >
          {SORT_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>
    </div>
  );
};

export const DynamicFilterBar = BoardFilterBar;
