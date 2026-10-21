import React, { useState } from 'react';
import { ChevronDown, Plus, CornerDownRight } from 'lucide-react';
import { Checkbox, Input } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import {
  useSubtasksQuery,
  useCreateSubtaskMutation,
  useToggleTaskMutation,
  formatTaskNumber,
} from '@/entities/task';

export interface TaskSubtasksSectionProps {
  workspaceId?: string;
  parentTaskId: string;
  onOpenTask?: (taskId: string) => void;
  onToggleTask?: (taskId: string) => void;
}

export const TaskSubtasksSection: React.FC<TaskSubtasksSectionProps> = ({
  workspaceId,
  parentTaskId,
  onOpenTask,
  onToggleTask,
}) => {
  const [isExpanded, setIsExpanded] = useState(true);
  const [newTitle, setNewTitle] = useState('');
  const [isInputFocused, setIsInputFocused] = useState(false);

  const { data: subtasks = [], isLoading } = useSubtasksQuery(workspaceId, parentTaskId);
  const createSubtaskMutation = useCreateSubtaskMutation(workspaceId);
  const toggleMutation = useToggleTaskMutation();

  const completedCount = subtasks.filter((s) => s.isCompleted).length;
  const totalCount = subtasks.length;

  const handleToggleSubtask = (subtaskId: string) => {
    if (onToggleTask) {
      onToggleTask(subtaskId);
    } else {
      toggleMutation.mutate(subtaskId);
    }
  };

  const handleCreateSubtask = async (e?: React.FormEvent) => {
    e?.preventDefault();
    const trimmed = newTitle.trim();
    if (!trimmed || createSubtaskMutation.isPending) return;

    try {
      await createSubtaskMutation.mutateAsync({
        title: trimmed,
        parentId: parentTaskId,
      });
      setNewTitle('');
    } catch {
      // Handled in mutation
    }
  };

  return (
    <div className="space-y-2 pt-2 border-t border-subtle/50">
      {/* Section Header */}
      <button
        type="button"
        onClick={() => setIsExpanded((prev) => !prev)}
        className="w-full flex items-center justify-between text-left group py-1 select-none"
      >
        <div className="flex items-center gap-1.5 text-[11px] font-mono uppercase tracking-wider text-muted group-hover:text-primary transition-colors">
          <ChevronDown
            className={cn(
              'w-3.5 h-3.5 transition-transform duration-200',
              !isExpanded && '-rotate-90'
            )}
          />
          <span>Подзадачи</span>
          {totalCount > 0 && (
            <span className="text-[10px] font-mono text-muted bg-canvas/60 px-1.5 py-0.2 rounded border border-subtle/60">
              {completedCount}/{totalCount}
            </span>
          )}
        </div>

        {totalCount > 0 && (
          <div className="w-20 bg-canvas h-1.5 rounded-full overflow-hidden border border-subtle/40">
            <div
              className="bg-accent-solid h-full rounded-full transition-all duration-300"
              style={{
                width: `${totalCount > 0 ? (completedCount / totalCount) * 100 : 0}%`,
              }}
            />
          </div>
        )}
      </button>

      {/* Subtasks List */}
      {isExpanded && (
        <div className="space-y-1.5 pl-1">
          {isLoading && subtasks.length === 0 && (
            <div className="py-2 text-xs text-muted/60 italic font-serif">Загрузка подзадач...</div>
          )}

          {subtasks.map((subtask) => {
            const shortNum = formatTaskNumber(subtask.id);
            return (
              <div
                key={subtask.id}
                className="group flex items-center gap-2.5 p-1.5 rounded-lg hover:bg-canvas/50 transition-colors"
              >
                <div
                  className="shrink-0"
                  onClick={(e) => e.stopPropagation()}
                >
                  <Checkbox
                    checked={subtask.isCompleted}
                    onChange={() => handleToggleSubtask(subtask.id)}
                    aria-label={`Завершить ${subtask.title}`}
                  />
                </div>

                <button
                  type="button"
                  onClick={() => onOpenTask?.(subtask.id)}
                  className="flex-1 text-left flex items-center justify-between gap-2 min-w-0"
                  title="Открыть подзадачу"
                >
                  <span
                    className={cn(
                      'text-xs truncate transition-colors',
                      subtask.isCompleted
                        ? 'line-through text-muted'
                        : 'text-primary group-hover:text-accent-solid'
                    )}
                  >
                    {subtask.title}
                  </span>
                  <span className="font-mono text-[10px] text-muted/60 shrink-0">
                    {shortNum}
                  </span>
                </button>
              </div>
            );
          })}

          {/* Inline Input for New Subtask */}
          <form
            onSubmit={handleCreateSubtask}
            className={cn(
              'flex items-center gap-2 px-2 py-1.5 rounded-lg border transition-all mt-1',
              isInputFocused
                ? 'bg-surface border-accent-solid/60 shadow-washi-sm'
                : 'bg-canvas/30 border-subtle/60 hover:border-subtle hover:bg-canvas/50'
            )}
          >
            <Plus className="w-3.5 h-3.5 text-muted shrink-0" />
            <Input
              type="text"
              value={newTitle}
              onChange={(e) => setNewTitle(e.target.value)}
              onFocus={() => setIsInputFocused(true)}
              onBlur={() => setIsInputFocused(false)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  handleCreateSubtask();
                }
              }}
              placeholder="Новая задача... (Enter для добавления)"
              disabled={createSubtaskMutation.isPending}
              className="border-0 bg-transparent p-0 text-xs text-primary placeholder:text-muted/60 focus:ring-0 focus:outline-none shadow-none h-auto"
            />
            {newTitle.trim() && (
              <button
                type="submit"
                disabled={createSubtaskMutation.isPending}
                className="shrink-0 text-[10px] font-mono text-accent-solid hover:text-accent-solid-hover px-1.5 py-0.5 rounded bg-accent-surface/50 border border-accent-solid/20"
              >
                <CornerDownRight className="w-3 h-3" />
              </button>
            )}
          </form>
        </div>
      )}
    </div>
  );
};
