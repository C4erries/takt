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
import type { Task, Priority, Tag } from '@/shared/api';
import {
  useUpdateTaskMutation,
  useSetTaskFocusMutation,
  useWorkspaceTagsQuery,
  useCreateTagMutation,
  useAddTagToTaskMutation,
  useRemoveTagFromTaskMutation,
} from '@/entities/task';
import { useProjectsQuery } from '@/entities/project';
import { useTaskStore } from '@/features/task-navigation';
import { MarkdownNotes } from './MarkdownNotes';

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
  { value: 'urgent', label: 'Срочно' },
];

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
  const { data: allTags = [] } = useWorkspaceTagsQuery(workspaceId);
  const createTagMutation = useCreateTagMutation(workspaceId);
  const addTagMutation = useAddTagToTaskMutation(workspaceId);
  const removeTagMutation = useRemoveTagFromTaskMutation(workspaceId);
  const { data: projects = [] } = useProjectsQuery(workspaceId);

  const [title, setTitle] = useState(task?.title ?? '');
  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [isAddingTag, setIsAddingTag] = useState(false);
  const [newTagName, setNewTagName] = useState('');

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

  const handleCreateAndAttachTag = async (e: React.FormEvent) => {
    e.preventDefault();
    const clean = newTagName.trim().replace(/^@/, '');
    if (!clean) return;

    const existing = allTags.find((t) => t.name.toLowerCase() === clean.toLowerCase());
    if (existing) {
      await addTagMutation.mutateAsync({ taskId: task.id, tag: existing });
    } else {
      const created = await createTagMutation.mutateAsync({ name: clean, colorHex: 'matcha' });
      await addTagMutation.mutateAsync({ taskId: task.id, tag: created });
    }
    setNewTagName('');
    setIsAddingTag(false);
  };

  return (
    <div className="w-80 shrink-0 bg-surface border border-subtle rounded-2xl p-4 shadow-washi flex flex-col justify-between overflow-y-auto max-h-[580px] select-none space-y-4">
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
              const active = task.priority === p.value;
              return (
                <button
                  key={p.value}
                  type="button"
                  onClick={() => updateTaskMutation.mutate({ taskId: task.id, data: { priority: p.value } })}
                  className={`text-[11px] px-2 py-0.5 rounded-full border transition-all ${
                    active
                      ? 'bg-accent-surface border-accent-solid/50 font-medium text-primary shadow-washi-sm'
                      : 'border-subtle/60 text-muted hover:text-primary hover:bg-canvas/50'
                  }`}
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
            <span className="text-[10px] font-mono uppercase tracking-wider text-muted">
              Теги контекста
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

          <div className="flex items-center gap-1 flex-wrap">
            {task.tags && task.tags.length > 0 ? (
              task.tags.map((tag: Tag) => (
                <span
                  key={tag.id}
                  className="inline-flex items-center gap-1 bg-accent-surface text-primary text-[11px] font-medium px-2 py-0.5 rounded-full border border-subtle"
                >
                  #{tag.name}
                  <button
                    type="button"
                    onClick={() => removeTagMutation.mutate({ taskId: task.id, tag })}
                    className="hover:text-muted"
                  >
                    <X className="w-2.5 h-2.5" />
                  </button>
                </span>
              ))
            ) : (
              <span className="text-[11px] text-muted italic">Нет тегов</span>
            )}
          </div>

          {isAddingTag && (
            <form onSubmit={handleCreateAndAttachTag} className="pt-1 flex items-center gap-1">
              <Input
                autoFocus
                placeholder="@контекст..."
                value={newTagName}
                onChange={(e) => setNewTagName(e.target.value)}
                className="text-xs bg-canvas/40 border-subtle h-7"
              />
              <Button type="submit" size="sm" variant="solid" className="h-7 text-xs px-2">
                Ок
              </Button>
            </form>
          )}
        </div>

        {/* Markdown Notes with Split-Toggle */}
        <MarkdownNotes
          initialValue={task.descriptionMarkdown ?? ''}
          onSave={(val) =>
            updateTaskMutation.mutate({ taskId: task.id, data: { descriptionMarkdown: val } })
          }
        />
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
