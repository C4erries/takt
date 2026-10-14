import React, { useState } from 'react';
import { Plus, Inbox } from 'lucide-react';
import { Button, Input, Skeleton, EmptyState, ErrorBanner } from '@/shared/components/ui';
import { TaskRow } from './TaskRow';
import type { Task, CreateInboxTaskRequest } from '@/shared/api';

export interface MainSheetProps {
  tasks: Task[];
  isLoading: boolean;
  error: string | null;
  onRetry: () => void;
  selectedIndex: number;
  onSelectTask: (index: number) => void;
  onToggleTask: (taskId: string) => void;
  onDeleteTask: (taskId: string) => void;
  onCreateTask: (req: CreateInboxTaskRequest) => Promise<void>;
  onOpenQuickCapture: () => void;
}

export const MainSheet: React.FC<MainSheetProps> = ({
  tasks,
  isLoading,
  error,
  onRetry,
  selectedIndex,
  onSelectTask,
  onToggleTask,
  onDeleteTask,
  onCreateTask,
  onOpenQuickCapture,
}) => {
  const [inlineTitle, setInlineTitle] = useState('');
  const [isSubmittingInline, setIsSubmittingInline] = useState(false);

  const handleInlineSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = inlineTitle.trim();
    if (!trimmed || isSubmittingInline) return;

    setIsSubmittingInline(true);
    try {
      await onCreateTask({ title: trimmed });
      setInlineTitle('');
    } finally {
      setIsSubmittingInline(false);
    }
  };

  return (
    <div className="flex-1 max-w-[880px] w-full bg-surface border border-subtle rounded-sheet shadow-washi flex flex-col overflow-hidden my-3">
      {/* Sheet Header */}
      <div className="p-8 pb-4 border-b border-subtle/50 space-y-1">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 text-xs font-mono uppercase tracking-wider text-muted">
            <span className="w-2 h-2 rounded-full bg-accent-solid inline-block" />
            <span>Horizon · Inbox</span>
          </div>
          <div className="text-xs font-mono text-muted">
            {tasks.length} {tasks.length === 1 ? 'задача' : 'задач'}
          </div>
        </div>

        <h1 className="font-serif text-3xl font-normal tracking-tight text-primary">
          Входящие
        </h1>
        <p className="text-xs text-muted">
          Мгновенный сброс мыслей, организация и разбор входящих задач
        </p>

        {/* Inline Quick Capture row */}
        <form onSubmit={handleInlineSubmit} className="pt-4 flex items-center gap-2">
          <Input
            value={inlineTitle}
            onChange={(e) => setInlineTitle(e.target.value)}
            placeholder="Записать мысль или задачу... (Enter для сохранения)"
            className="bg-canvas/50 border-subtle focus-within:bg-surface text-sm"
            disabled={isSubmittingInline}
          />
          <Button
            type="submit"
            variant="solid"
            size="md"
            isLoading={isSubmittingInline}
            disabled={!inlineTitle.trim()}
            leftIcon={<Plus className="w-4 h-4" />}
          >
            Добавить
          </Button>
        </form>
      </div>

      {/* Main Content Area: 4 States */}
      <div className="flex-1 overflow-y-auto p-6 space-y-1">
        {/* State 1: Loading (Skeleton) */}
        {isLoading && (
          <div className="space-y-3 py-2">
            {[1, 2, 3, 4, 5].map((i) => (
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
              title="Не удалось загрузить задачи"
              message={error}
              onRetry={onRetry}
            />
          </div>
        )}

        {/* State 3: Empty State */}
        {!isLoading && !error && tasks.length === 0 && (
          <div className="py-12">
            <EmptyState
              icon={<Inbox className="w-7 h-7 stroke-[1.5]" />}
              title="Входящие пусты"
              description="Нажмите C чтобы быстро записать мысль или воспользуйтесь полем ввода выше."
              action={
                <Button
                  variant="outline"
                  size="sm"
                  onClick={onOpenQuickCapture}
                  leftIcon={<Plus className="w-3.5 h-3.5" />}
                >
                  Новая задача (C)
                </Button>
              }
            />
          </div>
        )}

        {/* State 4: Success - Lined task list */}
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

      {/* Sheet Footer: Shortcuts indicator */}
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
