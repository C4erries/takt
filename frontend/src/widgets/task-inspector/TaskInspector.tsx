import React, { useState, useEffect } from 'react';
import {
  Calendar,
  Tag as TagIcon,
  CheckCircle2,
  Circle,
  X,
  Star,
  Plus,
  FolderClosed,
} from 'lucide-react';
import { Button, Input } from '@/shared/components/ui';
import type { Task, Priority } from '@/shared/api';
import {
  useUpdateTaskMutation,
  useSetTaskFocusMutation,
} from '@/entities/task';
import { useProjectsQuery } from '@/entities/project';
import { useUploadAttachmentMutation } from '@/entities/attachment';
import { useTaskStore } from '@/features/task-navigation';
import { TaskTagSection } from '@/widgets/task-drawer';
import { cn } from '@/shared/lib/cn';
import { MarkdownNotes } from './MarkdownNotes';
import { TaskAttachments } from './TaskAttachments';

export interface TaskInspectorProps {
  task: Task | null;
  workspaceId?: string;
  onClose?: () => void;
  onToggleTask?: (taskId: string) => void;
  onDeleteTask?: (taskId: string) => void;
}

const PRIORITIES: { value: Priority; label: string }[] = [
  { value: 'none', label: 'Обычный' },
  { value: 'low', label: 'Низкий' },
  { value: 'medium', label: 'Средний' },
  { value: 'high', label: 'Высокий' },
  { value: 'pending', label: 'В ожидании' },
];

const PRIORITY_ACTIVE_STYLES: Record<string, string> = {
  pending: 'bg-[#4A6B82]/15 text-[#4A6B82] border-[#4A6B82]/40',
  urgent: 'bg-[#4A6B82]/15 text-[#4A6B82] border-[#4A6B82]/40',
  high: 'bg-[#D97746]/15 text-[#D97746] border-[#D97746]/40',
  medium: 'bg-[#B8943F]/15 text-[#B8943F] border-[#B8943F]/40',
  low: 'bg-[#566B44]/15 text-[#566B44] border-[#566B44]/40',
  none: 'bg-surface text-primary border-subtle',
};

export const TaskInspector: React.FC<TaskInspectorProps> = ({
  task,
  workspaceId,
  onClose,
  onToggleTask,
  onDeleteTask,
}) => {
  const { setSelectedTaskId } = useTaskStore();
  const updateTaskMutation = useUpdateTaskMutation(workspaceId);
  const setFocusMutation = useSetTaskFocusMutation(workspaceId);
  const { data: projects = [] } = useProjectsQuery(workspaceId);
  const uploadAttachmentMutation = useUploadAttachmentMutation(workspaceId, task?.id);

  const [title, setTitle] = useState(task?.title ?? '');
  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [isAddingTag, setIsAddingTag] = useState(false);
  const [isDragOver, setIsDragOver] = useState(false);

  const handleDragOver = (e: React.DragEvent) => {
    if (e.dataTransfer.types.includes('Files')) {
      e.preventDefault();
      setIsDragOver(true);
    }
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
  };

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
    if (task && e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      for (const file of Array.from(e.dataTransfer.files)) {
        try {
          await uploadAttachmentMutation.mutateAsync(file);
        } catch {
          // ignore
        }
      }
    }
  };

  const handleUploadImageForNotes = async (file: File): Promise<string | null> => {
    if (!task) return null;
    try {
      const att = await uploadAttachmentMutation.mutateAsync(file);
      return att.contentUrl;
    } catch {
      return null;
    }
  };

  useEffect(() => {
    setTitle(task?.title ?? '');
    setIsEditingTitle(false);
    setIsAddingTag(false);
  }, [task?.id, task?.title]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !isEditingTitle && !isAddingTag) {
        setSelectedTaskId(null);
        onClose?.();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isEditingTitle, isAddingTag, setSelectedTaskId, onClose]);

  if (!task) {
    return (
      <div className="w-80 shrink-0 bg-surface border border-subtle rounded-2xl p-6 shadow-washi flex flex-col items-center justify-center text-center space-y-2 select-none h-[420px]">
        <div className="w-10 h-10 rounded-full bg-accent-surface/40 flex items-center justify-center text-muted border border-subtle">
          <Circle className="w-5 h-5" />
        </div>
        <h4 className="font-serif text-lg font-normal text-primary">Инспектор задачи</h4>
        <p className="text-xs text-muted max-w-[200px]">
          Используйте <kbd className="font-mono text-[10px] bg-canvas px-1 py-0.5 rounded border border-subtle">J</kbd> и <kbd className="font-mono text-[10px] bg-canvas px-1 py-0.5 rounded border border-subtle">K</kbd> для навигации или кликните строку.
        </p>
      </div>
    );
  }

  const isFocused = task.focusOrder !== null && task.focusOrder !== undefined;

  const handleTitleBlur = () => {
    setIsEditingTitle(false);
    const trimmed = title.trim();
    if (trimmed && trimmed !== task.title) {
      updateTaskMutation.mutate({ taskId: task.id, data: { title: trimmed } });
    }
  };

  return (
    <div
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      className={`w-80 shrink-0 bg-surface border border-subtle rounded-2xl p-4 shadow-washi flex flex-col justify-between overflow-y-auto max-h-[580px] select-none space-y-4 transition-colors ${
        isDragOver ? 'ring-2 ring-accent-solid/50 bg-accent-surface/10' : ''
      }`}
    >
      <div className="space-y-4">
        {/* Header bar */}
        <div className="flex items-center justify-between border-b border-subtle/50 pb-2.5">
          <button
            type="button"
            onClick={() =>
              setFocusMutation.mutate({
                taskId: task.id,
                focusOrder: isFocused ? null : 1,
              })
            }
            className={`flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-medium transition-colors ${
              isFocused
                ? 'bg-accent-solid text-surface'
                : 'bg-canvas text-muted hover:text-primary hover:bg-canvas/80'
            }`}
            title={isFocused ? 'Снять из фокуса дня' : 'Добавить в фокус дня'}
          >
            <Star className="w-3 h-3 fill-current" />
            <span>{isFocused ? 'В фокусе' : 'В фокус'}</span>
          </button>

          <button
            type="button"
            onClick={() => {
              setSelectedTaskId(null);
              onClose?.();
            }}
            className="p-1 rounded-md text-muted hover:text-primary hover:bg-canvas transition-colors"
            title="Закрыть (Esc)"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        </div>

        {/* Title */}
        <div>
          {isEditingTitle ? (
            <Input
              autoFocus
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              onBlur={handleTitleBlur}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleTitleBlur();
                if (e.key === 'Escape') {
                  setTitle(task.title);
                  setIsEditingTitle(false);
                }
              }}
              className="text-base font-serif bg-canvas/40 border-subtle"
            />
          ) : (
            <h3
              onClick={() => setIsEditingTitle(true)}
              className="font-serif text-lg font-normal text-primary leading-snug cursor-text hover:bg-canvas/30 rounded p-1 -m-1 transition-colors"
              title="Нажмите для редактирования"
            >
              {task.title}
            </h3>
          )}
        </div>

        {/* Priority row */}
        <div className="space-y-1.5">
          <label className="text-[10px] font-mono uppercase tracking-wider text-muted flex items-center gap-1.5">
            <TagIcon className="w-3 h-3" /> Приоритет
          </label>
          <div className="flex items-center gap-1 flex-wrap">
            {PRIORITIES.map((p) => {
              const active =
                task.priority === p.value ||
                (p.value === 'pending' && task.priority === 'urgent');
              return (
                <button
                  key={p.value}
                  type="button"
                  onClick={() => updateTaskMutation.mutate({ taskId: task.id, data: { priority: p.value } })}
                  className={cn(
                    'text-[11px] px-2 py-0.5 rounded-full border transition-all',
                    active
                      ? `${PRIORITY_ACTIVE_STYLES[p.value]} font-medium shadow-washi-sm`
                      : 'border-subtle/60 text-muted hover:text-primary hover:bg-canvas/50'
                  )}
                >
                  {p.label}
                </button>
              );
            })}
          </div>
        </div>

        {/* Due Date & Project row */}
        <div className="grid grid-cols-2 gap-2 text-xs">
          <div className="space-y-1">
            <label className="text-[10px] font-mono uppercase tracking-wider text-muted flex items-center gap-1">
              <Calendar className="w-3 h-3" /> Срок
            </label>
            <input
              type="date"
              value={task.dueDate ?? ''}
              onChange={(e) =>
                updateTaskMutation.mutate({ taskId: task.id, data: { dueDate: e.target.value || null } })
              }
              className="w-full text-xs font-mono bg-canvas/40 border border-subtle rounded-md px-2 py-1 text-primary focus:outline-none focus:border-accent-solid"
            />
          </div>

          <div className="space-y-1">
            <label className="text-[10px] font-mono uppercase tracking-wider text-muted flex items-center gap-1">
              <FolderClosed className="w-3 h-3" /> Проект
            </label>
            <select
              value={task.projectId ?? ''}
              onChange={(e) =>
                updateTaskMutation.mutate({ taskId: task.id, data: { projectId: e.target.value || null } })
              }
              className="w-full text-xs bg-canvas/40 border border-subtle rounded-md px-1.5 py-1 text-primary focus:outline-none focus:border-accent-solid"
            >
              <option value="">(Входящие)</option>
              {projects.map((proj) => (
                <option key={proj.id} value={proj.id}>
                  {proj.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Tags Section */}
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <span className="text-[10px] font-mono uppercase tracking-wider text-muted flex items-center gap-1.5">
              <TagIcon className="w-3 h-3" /> Теги контекста
            </span>
            <button
              type="button"
              onClick={() => setIsAddingTag((p) => !p)}
              className="text-muted hover:text-primary p-0.5"
              title="Добавить тег"
            >
              <Plus className="w-3 h-3" />
            </button>
          </div>

          <TaskTagSection
            taskId={task.id}
            tags={task.tags}
            workspaceId={workspaceId ?? ''}
            isAddingTag={isAddingTag}
            onToggleAddingTag={() => setIsAddingTag((p) => !p)}
          />
        </div>

        {/* Markdown Notes with Split-Toggle */}
        <MarkdownNotes
          initialValue={task.descriptionMarkdown ?? ''}
          onSave={(val) =>
            updateTaskMutation.mutate({ taskId: task.id, data: { descriptionMarkdown: val } })
          }
          onUploadImage={handleUploadImageForNotes}
        />

        {/* Attachments Section */}
        <TaskAttachments workspaceId={workspaceId} taskId={task.id} />
      </div>

      {/* Action buttons footer */}
      <div className="pt-3 border-t border-subtle/50 space-y-2">
        {onToggleTask && (
          <Button
            variant={task.isCompleted ? 'outline' : 'solid'}
            size="sm"
            className="w-full text-xs"
            onClick={() => onToggleTask(task.id)}
            leftIcon={task.isCompleted ? <Circle className="w-3.5 h-3.5" /> : <CheckCircle2 className="w-3.5 h-3.5" />}
          >
            {task.isCompleted ? 'Вернуть в работу' : 'Отметить выполненной'}
          </Button>
        )}
        {onDeleteTask && (
          <Button
            variant="ghost"
            size="sm"
            className="w-full text-xs text-muted hover:text-primary"
            onClick={() => {
              onDeleteTask(task.id);
              setSelectedTaskId(null);
            }}
          >
            Удалить задачу
          </Button>
        )}
      </div>
    </div>
  );
};
