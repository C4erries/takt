import React, { useState } from 'react';
import { useDroppable } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { Plus, X } from 'lucide-react';
import { Badge, Button, Input } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import type { Task } from '@/shared/api';
import type { Column, ColumnStatusCategory } from '@/entities/project';
import { KanbanCard } from './KanbanCard';

export interface KanbanColumnProps {
  column: Column;
  tasks: Task[];
  onToggleTask?: (taskId: string) => void;
  onSelectTask?: (taskId: string) => void;
  selectedTaskId?: string | null;
  onCreateTask?: (title: string, columnId: string) => Promise<void>;
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
  onToggleTask,
  onSelectTask,
  selectedTaskId,
  onCreateTask,
}) => {
  const { setNodeRef, isOver } = useDroppable({
    id: column.id,
    data: {
      type: 'Column',
      column,
    },
  });

  const [isAdding, setIsAdding] = useState(false);
  const [newTitle, setNewTitle] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

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
      setIsAdding(false);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div
      ref={setNodeRef}
      className={cn(
        'w-72 shrink-0 flex flex-col bg-canvas/60 border border-subtle rounded-xl p-3 transition-colors duration-150',
        isOver && 'border-accent-solid/50 bg-accent-surface/20'
      )}
    >
      {/* Column Header */}
      <div className="flex items-center justify-between px-1 pb-3 select-none">
        <div className="flex items-center gap-2 min-w-0">
          <span className={cn('w-2 h-2 rounded-full shrink-0', categoryConfig.dot)} />
          <h3 className="text-xs font-semibold tracking-wide uppercase text-primary truncate">
            {column.name}
          </h3>
          <Badge variant="subtle" size="sm" className="font-mono text-[10px] px-1.5 py-0">
            {tasks.length}
          </Badge>
        </div>

        <Button
          variant="ghost"
          size="sm"
          onClick={() => setIsAdding((prev) => !prev)}
          className="h-6 w-6 p-0 text-muted hover:text-primary"
          title="Добавить карточку"
          aria-label={`Добавить задачу в ${column.name}`}
        >
          {isAdding ? <X className="w-3.5 h-3.5" /> : <Plus className="w-3.5 h-3.5" />}
        </Button>
      </div>

      {/* Quick Add Form */}
      {isAdding && (
        <form onSubmit={handleQuickAdd} className="mb-3 space-y-2">
          <Input
            autoFocus
            value={newTitle}
            onChange={(e) => setNewTitle(e.target.value)}
            placeholder="Название карточки..."
            disabled={isSubmitting}
            className="text-xs bg-surface border-subtle"
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
              className="text-[11px] h-7 px-2"
            >
              Отмена
            </Button>
            <Button
              type="submit"
              variant="solid"
              size="sm"
              isLoading={isSubmitting}
              disabled={!newTitle.trim()}
              className="text-[11px] h-7 px-2.5"
            >
              Создать
            </Button>
          </div>
        </form>
      )}

      {/* Droppable Task List */}
      <SortableContext items={tasks.map((t) => t.id)} strategy={verticalListSortingStrategy}>
        <div className="flex-1 min-h-[140px] space-y-2 overflow-y-auto pr-0.5">
          {tasks.map((task) => (
            <KanbanCard
              key={task.id}
              task={task}
              isSelected={selectedTaskId === task.id}
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
