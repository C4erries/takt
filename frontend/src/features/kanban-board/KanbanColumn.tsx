import React, { useState, useRef, useEffect } from 'react';
import { useSortable, SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Plus, MoreHorizontal, ArrowRight, GripVertical } from 'lucide-react';
import { Badge, Button } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import type { Task } from '@/shared/api';
import {
  useProjectBoardsQuery,
  useMoveColumnMutation,
  type Column,
  type ColumnStatusCategory,
} from '@/entities/project';
import { KanbanCard } from './KanbanCard';

export interface KanbanColumnProps {
  column: Column;
  tasks: Task[];
  projectId?: string;
  workspaceId?: string;
  onToggleTask?: (taskId: string) => void;
  onSelectTask?: (taskId: string) => void;
  selectedTaskId?: string | null;
  onCreateTask?: (title: string, columnId: string) => Promise<void>;
  isOverlay?: boolean;
}

const CATEGORY_STYLES: Record<ColumnStatusCategory, { dot: string; label: string }> = {
  backlog: { dot: 'bg-muted', label: 'Бэклог' },
  todo: { dot: 'bg-accent-solid', label: 'К выполнению' },
  in_progress: { dot: 'bg-tag-persimmon', label: 'В работе' },
  done: { dot: 'bg-accent-solid', label: 'Готово' },
};

export const KanbanColumn: React.FC<KanbanColumnProps> = ({
  column,
  tasks,
  projectId,
  workspaceId,
  onToggleTask,
  onSelectTask,
  selectedTaskId,
  onCreateTask,
  isOverlay = false,
}) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
    isOver,
  } = useSortable({
    id: column.id,
    data: {
      type: 'Column',
      column,
    },
    disabled: isOverlay,
  });

  const style: React.CSSProperties = {
    transform: CSS.Translate.toString(transform),
    transition,
  };

  const [isAdding, setIsAdding] = useState(false);
  const [newTitle, setNewTitle] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const boardsQuery = useProjectBoardsQuery(workspaceId, projectId);
  const moveColumnMutation = useMoveColumnMutation(workspaceId, column.boardId);
  const otherBoards = (boardsQuery.data ?? []).filter((b) => b.id !== column.boardId);

  // Close dropdown menu on click outside
  useEffect(() => {
    if (!isMenuOpen) return;
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setIsMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isMenuOpen]);

  // Autofocus when adding
  useEffect(() => {
    if (isAdding) {
      inputRef.current?.focus();
    }
  }, [isAdding]);

  const categoryConfig = CATEGORY_STYLES[column.statusCategory] || CATEGORY_STYLES.todo;

  const handleQuickAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = newTitle.trim();
    if (!trimmed || isSubmitting) return;

    setIsSubmitting(true);
    try {
      if (onCreateTask) {
        await onCreateTask(trimmed, column.id);
      }
      setNewTitle('');
      inputRef.current?.focus();
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        'w-72 shrink-0 flex flex-col bg-canvas/60 border border-subtle rounded-xl p-3 transition-colors duration-150',
        isOver && !isDragging && 'border-accent-solid/50 bg-accent-surface/20',
        isDragging && 'opacity-30 border-dashed border-subtle/80 shadow-none',
        isOverlay && 'shadow-washi-modal ring-1 ring-accent-solid/40 rotate-1 cursor-grabbing bg-canvas/90'
      )}
    >
      {/* Column Header */}
      <div className="flex items-center justify-between px-1 pb-2.5 select-none">
        <div
          {...(!isOverlay ? { ...attributes, ...listeners } : {})}
          className="flex items-center gap-2 min-w-0 cursor-grab active:cursor-grabbing flex-1 mr-1"
          title={`Перетащить колонку ${column.name}`}
        >
          {!isOverlay && (
            <div
              className="text-muted/40 hover:text-muted p-0.5 -ml-1 rounded transition-colors shrink-0"
              title="Перетащить колонку"
              aria-label={`Перетащить колонку ${column.name}`}
            >
              <GripVertical className="w-3.5 h-3.5" />
            </div>
          )}
          <span className={cn('w-2 h-2 rounded-full shrink-0', categoryConfig.dot)} />
          <h3 className="text-xs font-semibold tracking-wide uppercase text-primary truncate">
            {column.name}
          </h3>
          <Badge variant="subtle" size="sm" className="font-mono text-[10px] px-1.5 py-0">
            {tasks.length}
          </Badge>
        </div>

        <div className="flex items-center gap-1" onPointerDown={(e) => e.stopPropagation()}>
          <div className="relative" ref={menuRef}>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsMenuOpen((prev) => !prev)}
              className={cn(
                'h-6 w-6 p-0 text-muted hover:text-primary',
                isMenuOpen && 'bg-canvas/60 text-primary'
              )}
              title="Меню колонки"
              aria-label={`Меню колонки ${column.name}`}
            >
              <MoreHorizontal className="w-3.5 h-3.5" />
            </Button>

            {isMenuOpen && (
              <div className="absolute right-0 top-full mt-1 w-52 z-30 bg-surface border border-subtle rounded-xl shadow-washi p-1.5 space-y-1 text-xs animate-in fade-in zoom-in-95 duration-100">
                <div className="px-2 py-1 text-[10px] font-semibold text-muted tracking-wider uppercase">
                  Переместить на доску
                </div>

                {otherBoards.length === 0 ? (
                  <div className="px-2 py-1.5 text-[11px] text-muted italic">
                    Нет других досок
                  </div>
                ) : (
                  <div className="space-y-0.5">
                    {otherBoards.map((targetBoard) => (
                      <button
                        key={targetBoard.id}
                        type="button"
                        onClick={() => {
                          setIsMenuOpen(false);
                          moveColumnMutation.mutate({
                            columnId: column.id,
                            targetBoardId: targetBoard.id,
                          });
                        }}
                        className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors group"
                      >
                        <span className="truncate">{targetBoard.name}</span>
                        <ArrowRight className="w-3 h-3 text-muted group-hover:text-primary shrink-0 ml-1.5" />
                      </button>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Inline Task Add row under Column Header */}
      {!isOverlay && (
        <div className="mb-2">
          {isAdding ? (
            <form
              onSubmit={handleQuickAdd}
              className="p-2 rounded-xl bg-surface border border-subtle shadow-washi space-y-2 animate-in fade-in duration-100"
            >
              <input
                ref={inputRef}
                type="text"
                autoFocus
                value={newTitle}
                onChange={(e) => setNewTitle(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Escape') {
                    setIsAdding(false);
                    setNewTitle('');
                  }
                }}
                placeholder="Название задачи... (Enter)"
                disabled={isSubmitting}
                className="w-full px-2.5 py-1.5 text-xs bg-canvas/40 border border-subtle rounded-md text-primary placeholder:text-muted/60 focus:outline-none focus:border-accent-solid"
              />
              <div className="flex items-center justify-end gap-1.5">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setIsAdding(false);
                    setNewTitle('');
                  }}
                  className="text-[11px] h-6 px-2 text-muted hover:text-primary"
                >
                  Отмена
                </Button>
                <Button
                  type="submit"
                  variant="solid"
                  size="sm"
                  isLoading={isSubmitting}
                  disabled={!newTitle.trim()}
                  className="text-[11px] h-6 px-2.5"
                >
                  Добавить
                </Button>
              </div>
            </form>
          ) : (
            <button
              type="button"
              onClick={() => setIsAdding(true)}
              className="w-full text-xs text-muted hover:text-primary hover:bg-surface/80 rounded-lg py-1.5 px-2.5 flex items-center gap-1.5 transition-colors border border-transparent hover:border-subtle/50 select-none"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>Добавить задачу</span>
            </button>
          )}
        </div>
      )}

      {/* Droppable Task List */}
      <SortableContext items={tasks.map((t) => t.id)} strategy={verticalListSortingStrategy}>
        <div className="flex-1 min-h-[140px] space-y-2 overflow-y-auto pr-0.5">
          {tasks.map((task) => (
            <KanbanCard
              key={task.id}
              task={task}
              isSelected={selectedTaskId === task.id}
              isOverlay={isOverlay}
              onSelect={onSelectTask}
              onToggle={onToggleTask}
            />
          ))}

          {tasks.length === 0 && !isAdding && (
            <div
              className={cn(
                'h-24 rounded-lg border border-dashed border-subtle/70 flex items-center justify-center text-center p-3 select-none',
                isOver ? 'border-accent-solid/60 bg-accent-surface/30' : 'bg-surface/30'
              )}
            >
              <p className="text-[11px] text-muted">Перетащите задачу сюда</p>
            </div>
          )}
        </div>
      </SortableContext>
    </div>
  );
};
