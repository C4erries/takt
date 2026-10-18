import React from 'react';
import { Trash2 } from 'lucide-react';
import { Checkbox, Badge } from '@/shared/components/ui';
import type { Task, Priority } from '@/shared/api';

export interface TaskRowProps {
  task: Task;
  isSelected: boolean;
  onSelect: () => void;
  onToggle: (taskId: string) => void;
  onDelete?: (taskId: string) => void;
}

const getPriorityBadge = (priority: Priority) => {
  switch (priority) {
    case 'pending':
    case 'urgent':
      return <Badge variant="denim" size="sm">В ожидании</Badge>;
    case 'high':
      return <Badge variant="persimmon" size="sm">Высокий</Badge>;
    case 'medium':
      return <Badge variant="straw" size="sm">Средний</Badge>;
    case 'low':
      return <Badge variant="subtle" size="sm">Низкий</Badge>;
    case 'none':
    default:
      return null;
  }
};

export const TaskRow: React.FC<TaskRowProps> = ({
  task,
  isSelected,
  onSelect,
  onToggle,
  onDelete,
}) => {
  return (
    <div
      role="row"
      aria-selected={isSelected}
      onClick={onSelect}
      className={`group relative flex items-center justify-between min-h-[44px] px-3 py-2 rounded-lg border-b border-subtle/50 transition-all duration-150 ease-spring-jk cursor-pointer select-none ${
        isSelected
          ? 'bg-accent-surface shadow-washi-sm'
          : 'hover:bg-canvas/50'
      }`}
    >
      {/* Left: Checkbox & Title */}
      <div className="flex items-center gap-3 min-w-0 flex-1">
        <div
          onClick={(e) => {
            e.stopPropagation();
          }}
          className="shrink-0 flex items-center"
        >
          <Checkbox
            checked={task.isCompleted}
            onChange={() => onToggle(task.id)}
            aria-label={`Отметить задачу ${task.title}`}
          />
        </div>

        <span
          className={`text-sm truncate font-sans transition-colors ${
            task.isCompleted
              ? 'line-through text-muted'
              : 'text-primary'
          }`}
        >
          {task.title}
        </span>
      </div>

      {/* Right: Priority Badge & Action Button */}
      <div className="flex items-center gap-2 shrink-0 ml-3">
        {getPriorityBadge(task.priority)}

        {onDelete && (
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onDelete(task.id);
            }}
            title="Удалить задачу"
            className="opacity-0 group-hover:opacity-100 p-1 rounded-md text-muted hover:text-primary hover:bg-canvas transition-all"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        )}
      </div>
    </div>
  );
};
