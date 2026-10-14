import React, { useState } from 'react';
import { Plus, Zap, Sun } from 'lucide-react';
import { Button, Input, Skeleton, EmptyState, ErrorBanner } from '@/shared/components/ui';
import { TaskRow } from './TaskRow';
import type { Task, CreateInboxTaskRequest } from '@/shared/api';

export interface TodayViewProps {
  tasks: Task[];
  isLoading: boolean;
  error: string | null;
  onRetry: () => void;
  selectedIndex: number;
  onSelectTask: (index: number) => void;
  onToggleTask: (taskId: string) => void;
  onDeleteTask: (taskId: string) => void;
  onCreateTodayTask: (req: CreateInboxTaskRequest) => Promise<void>;
  onOpenQuickCapture: () => void;
}

export const TodayView: React.FC<TodayViewProps> = ({
  tasks,
  isLoading,
  error,
  onRetry,
  selectedIndex,
  onSelectTask,
  onToggleTask,
  onDeleteTask,
  onCreateTodayTask,
  onOpenQuickCapture,
}) => {
  const [inlineTitle, setInlineTitle] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const today = new Date();
  const formattedDate = today.toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'long',
    weekday: 'long',
  });
  const todayIso = today.toISOString().split('T')[0];

  const handleInlineSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = inlineTitle.trim();
    if (!trimmed || isSubmitting) return;

    setIsSubmitting(true);
    try {
      await onCreateTodayTask({
        title: trimmed,
        dueDate: todayIso,
      });
      setInlineTitle('');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Header */}
      <div className="p-8 pb-4 border-b border-subtle/50 space-y-1">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 text-xs font-mono uppercase tracking-wider text-muted">
            <span className="w-2 h-2 rounded-full bg-accent-solid inline-block" />
            <span>Horizon · Сегодня</span>
          </div>
          <div className="text-xs font-mono text-muted">
            {tasks.length} {tasks.length === 1 ? 'задача' : 'задач'}
          </div>
        </div>

        <h1 className="font-serif text-3xl font-normal tracking-tight text-primary">
          Сегодня
        </h1>
        <p className="text-xs text-muted capitalize">
          {formattedDate} · Focus of the Day
        </p>

        {/* Inline Quick Capture row */}
        <form onSubmit={handleInlineSubmit} className="pt-4 flex items-center gap-2">
          <Input
            value={inlineTitle}
            onChange={(e) => setInlineTitle(e.target.value)}
            placeholder="Фокусная задача на сегодня... (Enter для сохранения)"
            className="bg-canvas/50 border-subtle focus-within:bg-surface text-sm"
            disabled={isSubmitting}
          />
          <Button
            type="submit"
            variant="solid"
            size="md"
            isLoading={isSubmitting}
            disabled={!inlineTitle.trim()}
            leftIcon={<Plus className="w-4 h-4" />}
          >
            В фокус
          </Button>
        </form>
      </div>

      {/* Content: 4 States */}
      <div className="flex-1 overflow-y-auto p-6 space-y-1">
        {/* State 1: Loading */}
        {isLoading && (
          <div className="space-y-3 py-2">
            {[1, 2, 3, 4].map((i) => (
              <div
                key={i}
                className="flex items-center gap-3 p-3 rounded-lg border-b border-subtle/40"
              >
                <Skeleton variant="circle" className="w-[18px] h-[18px] shrink-0" />
                <Skeleton variant="text" className="w-3/5 h-4" />
                <Skeleton variant="rect" className="w-14 h-5 ml-auto" />
              </div>
            ))}
          </div>
        )}

        {/* State 2: Error */}
        {!isLoading && error && (
          <div className="py-6">
            <ErrorBanner
              title="Не удалось загрузить задачи на сегодня"
              message={error}
              onRetry={onRetry}
            />
          </div>
        )}

        {/* State 3: Empty State */}
        {!isLoading && !error && tasks.length === 0 && (
          <div className="py-12">
            <EmptyState
              icon={<Sun className="w-7 h-7 stroke-[1.5]" />}
              title="Фокус дня свободен"
              description="На сегодня нет запланированных задач или все задачи завершены. Добавьте задачу в фокус через поле выше или нажмите C."
              action={
                <Button
                  variant="outline"
                  size="sm"
                  onClick={onOpenQuickCapture}
                  leftIcon={<Zap className="w-3.5 h-3.5" />}
                >
                  Новая мысль (C)
                </Button>
              }
            />
          </div>
        )}

        {/* State 4: Success */}
        {!isLoading && !error && tasks.length > 0 && (
          <div className="space-y-0.5">
            {tasks.map((task, index) => (
              <TaskRow
                key={task.id}
                task={task}
                isSelected={selectedIndex === index}
                onSelect={() => onSelectTask(index)}
                onToggle={onToggleTask}
                onDelete={onDeleteTask}
              />
            ))}
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="p-3 px-6 border-t border-subtle/40 bg-canvas/30 flex items-center justify-between text-[11px] font-mono text-muted">
        <div className="flex items-center gap-3">
          <span>
            <kbd className="px-1.5 py-0.5 rounded bg-surface border border-subtle mr-1 text-[10px]">
              J
            </kbd>
            <kbd className="px-1.5 py-0.5 rounded bg-surface border border-subtle mr-1 text-[10px]">
              K
            </kbd>
            навигация
          </span>
          <span>
            <kbd className="px-1.5 py-0.5 rounded bg-surface border border-subtle mr-1 text-[10px]">
              Space
            </kbd>
            завершить
          </span>
        </div>
        <div>
          <kbd className="px-1.5 py-0.5 rounded bg-surface border border-subtle mr-1 text-[10px]">
            C
          </kbd>
          захват
        </div>
      </div>
    </div>
  );
};
