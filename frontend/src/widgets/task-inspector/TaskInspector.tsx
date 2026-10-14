import React from 'react';
import { Calendar, Tag, CheckCircle2, Circle } from 'lucide-react';
import { Badge, Button } from '@/shared/components/ui';
import type { Task } from '@/shared/api';

export interface TaskInspectorProps {
  task: Task | null;
  onToggleTask?: (taskId: string) => void;
  onDeleteTask?: (taskId: string) => void;
}

export const TaskInspector: React.FC<TaskInspectorProps> = ({
  task,
  onToggleTask,
  onDeleteTask,
}) => {
  if (!task) {
    return (
      <div className="w-72 shrink-0 bg-surface border border-subtle rounded-2xl p-6 shadow-washi flex flex-col items-center justify-center text-center space-y-2 select-none">
        <div className="w-10 h-10 rounded-full bg-accent-surface/40 flex items-center justify-center text-muted border border-subtle">
          <Circle className="w-5 h-5" />
        </div>
        <h4 className="font-serif text-lg font-normal text-primary">Инспектор задачи</h4>
        <p className="text-xs text-muted max-w-[200px]">
          Используйте клавиши <kbd className="font-mono text-[10px] bg-canvas px-1 py-0.5 rounded border border-subtle">J</kbd> и <kbd className="font-mono text-[10px] bg-canvas px-1 py-0.5 rounded border border-subtle">K</kbd> для выбора строки задачи.
        </p>
      </div>
    );
  }

  const createdDate = new Date(task.createdAt).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <div className="w-72 shrink-0 bg-surface border border-subtle rounded-2xl p-5 shadow-washi flex flex-col justify-between space-y-4 select-none">
      <div className="space-y-4">
        {/* Header: Horizon context */}
        <div className="flex items-center justify-between border-b border-subtle/50 pb-3">
          <span className="text-[10px] font-mono uppercase tracking-wider text-muted">
            Детали задачи
          </span>
          <Badge
            variant={task.isCompleted ? 'matcha' : 'subtle'}
            size="sm"
            dot
          >
            {task.isCompleted ? 'Завершена' : 'В работе'}
          </Badge>
        </div>

        {/* Task Title */}
        <div>
          <h3 className="font-serif text-xl font-normal text-primary leading-snug">
            {task.title}
          </h3>
        </div>

        {/* Meta items */}
        <div className="space-y-2.5 text-xs">
          <div className="flex items-center justify-between text-muted">
            <span className="flex items-center gap-1.5">
              <Tag className="w-3.5 h-3.5" />
              Приоритет
            </span>
            <span className="font-medium text-primary capitalize">
              {task.priority !== 'none' ? task.priority : 'Обычный'}
            </span>
          </div>

          <div className="flex items-center justify-between text-muted">
            <span className="flex items-center gap-1.5">
              <Calendar className="w-3.5 h-3.5" />
              Создана
            </span>
            <span className="font-mono text-[11px] text-primary">{createdDate}</span>
          </div>
        </div>

        {/* Description section */}
        {task.descriptionMarkdown && (
          <div className="pt-2 border-t border-subtle/40">
            <div className="text-[10px] font-mono uppercase text-muted mb-1">
              Заметки
            </div>
            <p className="text-xs text-primary leading-relaxed bg-canvas/40 p-2.5 rounded-lg border border-subtle/50">
              {task.descriptionMarkdown}
            </p>
          </div>
        )}
      </div>

      {/* Action buttons */}
      <div className="pt-3 border-t border-subtle/50 space-y-2">
        {onToggleTask && (
          <Button
            variant={task.isCompleted ? 'outline' : 'solid'}
            size="sm"
            className="w-full text-xs"
            onClick={() => onToggleTask(task.id)}
            leftIcon={
              task.isCompleted ? (
                <Circle className="w-3.5 h-3.5" />
              ) : (
                <CheckCircle2 className="w-3.5 h-3.5" />
              )
            }
          >
            {task.isCompleted ? 'Вернуть в работу' : 'Отметить выполненной'}
          </Button>
        )}

        {onDeleteTask && (
          <Button
            variant="ghost"
            size="sm"
            className="w-full text-xs text-muted hover:text-primary"
            onClick={() => onDeleteTask(task.id)}
          >
            Удалить задачу
          </Button>
        )}
      </div>
    </div>
  );
};
