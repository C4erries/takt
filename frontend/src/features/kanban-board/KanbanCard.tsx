import React from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { GripVertical } from 'lucide-react';
import { Checkbox, Badge } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import type { Task, Priority } from '@/shared/api';

export interface KanbanCardProps {
  task: Task;
  onToggle?: (taskId: string) => void;
  onSelect?: (taskId: string) => void;
  isSelected?: boolean;
  isOverlay?: boolean;
}

const PRIORITY_BADGES: Record<Priority, { label: string; variant: 'persimmon' | 'denim' | 'straw' | 'matcha' | 'subtle' } | null> = {
  urgent: { label: 'Срочно', variant: 'persimmon' },
  high: { label: 'Высокий', variant: 'denim' },
  medium: { label: 'Средний', variant: 'straw' },
  low: { label: 'Низкий', variant: 'subtle' },
  none: null,
};

export const KanbanCard: React.FC<KanbanCardProps> = ({
  task,
  onToggle,
  onSelect,
  isSelected = false,
  isOverlay = false,
}) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: task.id,
    data: {
      type: 'Task',
      task,
    },
    disabled: isOverlay,
  });

  const style: React.CSSProperties = {
    transform: CSS.Translate.toString(transform),
    transition,
  };

  const priorityBadge = PRIORITY_BADGES[task.priority];

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...(!isOverlay ? { ...attributes, ...listeners } : {})}
      onClick={() => onSelect?.(task.id)}
      className={cn(
        'group relative bg-surface border border-subtle rounded-lg p-3 select-none transition-shadow duration-150',
        'hover:border-subtle hover:shadow-washi',
        isSelected && 'ring-1 ring-accent-solid/30 bg-accent-surface/30',
        isDragging && 'opacity-30 border-dashed border-subtle/80 shadow-none',
        isOverlay && 'shadow-washi-modal ring-1 ring-accent-solid/40 rotate-1 cursor-grabbing bg-surface',
        !isOverlay && 'cursor-grab active:cursor-grabbing'
      )}
    >
      <div className="flex items-start gap-2">
        {/* Drag handle */}
        <div
          className="mt-0.5 text-muted/40 group-hover:text-muted/70 p-0.5 -ml-1 rounded transition-colors shrink-0"
          title="Перетащить карточку"
        >
          <GripVertical className="w-3.5 h-3.5" />
        </div>

        {/* Task Details */}
        <div className="flex-1 min-w-0 space-y-1.5 pr-1">
          <p
            className={cn(
              'text-xs font-medium leading-snug break-words transition-all duration-150',
              task.isCompleted
                ? 'line-through text-muted opacity-70'
                : 'text-primary'
            )}
          >
            {task.title}
          </p>

          {task.descriptionMarkdown && (
            <p className="text-[11px] text-muted line-clamp-2 leading-relaxed">
              {task.descriptionMarkdown}
            </p>
          )}

          {/* Badges / Metadata */}
          <div className="flex items-center gap-1.5 pt-0.5 flex-wrap">
            {priorityBadge && (
              <Badge variant={priorityBadge.variant} size="sm">
                {priorityBadge.label}
              </Badge>
            )}
            {task.dueDate && (
              <span className="font-mono text-[10px] text-muted">
                {new Date(task.dueDate).toLocaleDateString('ru-RU', {
                  day: 'numeric',
                  month: 'short',
                })}
              </span>
            )}
          </div>
        </div>

        {/* Round 18px GTD Checkbox on top-right */}
        <div
          className="pt-0.5 shrink-0 ml-auto"
          onClick={(e) => e.stopPropagation()}
          onPointerDown={(e) => e.stopPropagation()}
        >
          <Checkbox
            checked={task.isCompleted}
            onChange={() => onToggle?.(task.id)}
            aria-label={`Завершить задачу ${task.title}`}
          />
        </div>
      </div>
    </div>
  );
};
