import React, { useState, useEffect, useRef } from 'react';
import {
  X,
  Star,
  CheckCircle2,
  Circle,
  Trash2,
  FolderClosed,
} from 'lucide-react';
import type { Priority, Tag } from '@/shared/api';
import {
  useUpdateTaskMutation,
  useSetTaskFocusMutation,
  useWorkspaceTagsQuery,
  useCreateTagMutation,
  useAddTagToTaskMutation,
  useRemoveTagFromTaskMutation,
} from '@/entities/task';
import {
  useProjectsQuery,
  useProjectBoardsQuery,
  useBoardColumnsQuery,
  useMoveTaskMutation,
} from '@/entities/project';
import { useUploadAttachmentMutation } from '@/entities/attachment';
import { MarkdownNotes, TaskAttachments } from '@/widgets/task-inspector';
import { Button, Input } from '@/shared/components/ui';
import { QuickActionBar } from './QuickActionBar';
import type { TaskDrawerProps } from './types';

const PRIORITIES: { value: Priority; label: string }[] = [
  { value: 'none', label: 'Обычный' }, { value: 'low', label: 'Низкий' },
  { value: 'medium', label: 'Средний' }, { value: 'high', label: 'Высокий' },
  { value: 'urgent', label: 'Срочно' },
];

export const TaskDrawer: React.FC<TaskDrawerProps> = ({
  task,
  workspaceId,
  onClose,
  onToggleTask,
  onDeleteTask,
}) => {
  const isOpen = Boolean(task);

  const updateTaskMutation = useUpdateTaskMutation(workspaceId);
  const setFocusMutation = useSetTaskFocusMutation(workspaceId);
  const { data: allTags = [] } = useWorkspaceTagsQuery(workspaceId);
  const createTagMutation = useCreateTagMutation(workspaceId);
  const addTagMutation = useAddTagToTaskMutation(workspaceId);
  const removeTagMutation = useRemoveTagFromTaskMutation(workspaceId);
  const { data: projects = [] } = useProjectsQuery(workspaceId);

  const boardsQuery = useProjectBoardsQuery(workspaceId, task?.projectId ?? undefined);
  const activeBoard = boardsQuery.data?.[0];
  const columnsQuery = useBoardColumnsQuery(workspaceId, activeBoard?.id);
  const columns = columnsQuery.data ?? [];
  const moveTaskMutation = useMoveTaskMutation(workspaceId, task?.projectId ?? undefined);

  const uploadAttachmentMutation = useUploadAttachmentMutation(workspaceId, task?.id);

  const [title, setTitle] = useState(task?.title ?? '');
  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [isAddingTag, setIsAddingTag] = useState(false);
  const [newTagName, setNewTagName] = useState('');

  const fileInputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    setTitle(task?.title ?? '');
    setIsEditingTitle(false);
    setIsAddingTag(false);
  }, [task?.id, task?.title]);

  // Закрытие по Escape
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen && !isEditingTitle && !isAddingTag) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, isEditingTitle, isAddingTag, onClose]);

  if (!task && !isOpen) return null;

  const isFocused = task ? task.focusOrder !== null && task.focusOrder !== undefined : false;

  const handleTitleBlur = () => {
    setIsEditingTitle(false);
    if (!task) return;
    const trimmed = title.trim();
    if (trimmed && trimmed !== task.title) {
      updateTaskMutation.mutate({ taskId: task.id, data: { title: trimmed } });
    }
  };

  const handleCreateAndAttachTag = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!task) return;
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

  const handleUploadImageForNotes = async (file: File): Promise<string | null> => {
    if (!task) return null;
    try {
      const att = await uploadAttachmentMutation.mutateAsync(file);
      return att.contentUrl;
    } catch {
      return null;
    }
  };

  const handleFileInputChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0 && task) {
      for (const file of Array.from(e.target.files)) {
        try {
          await uploadAttachmentMutation.mutateAsync(file);
        } catch {
          // Handled by react-query
        }
      }
      e.target.value = '';
    }
  };

  const handleColumnChange = (columnId: string) => {
    if (!task || !columnId) return;
    moveTaskMutation.mutate({
      taskId: task.id,
      data: {
        columnId,
        position: task.position,
        projectId: task.projectId ?? undefined,
      },
    });
  };

  return (
    <>
      {/* Mobile/Tablet Backdrop (modal on < lg, non-modal on >= lg) */}
      {isOpen && (
        <div
          className="fixed inset-0 bg-primary/20 z-40 lg:hidden transition-opacity"
          onClick={onClose}
        />
      )}

      {/* Slide-out Sheet 560px */}
      <aside
        className={`fixed top-0 right-0 h-full w-full max-w-[560px] bg-surface z-40 border-l border-subtle shadow-washi-modal flex flex-col justify-between transition-transform duration-300 ease-spring-jk ${
          isOpen ? 'translate-x-0' : 'translate-x-full pointer-events-none'
        }`}
      >
        <input
          type="file"
          ref={fileInputRef}
          onChange={handleFileInputChange}
          className="hidden"
          multiple
        />

        {task ? (
          <div className="flex-1 flex flex-col overflow-y-auto p-6 space-y-5">
            {/* Header top bar */}
            <div className="flex items-center justify-between border-b border-subtle/50 pb-3">
              <button
                type="button"
                onClick={() =>
                  setFocusMutation.mutate({
                    taskId: task.id,
                    focusOrder: isFocused ? null : 1,
                  })
                }
                className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium transition-colors ${
                  isFocused
                    ? 'bg-accent-solid text-surface'
                    : 'bg-canvas text-muted hover:text-primary hover:bg-canvas/80'
                }`}
                title={isFocused ? 'Снять из фокуса дня' : 'Добавить в фокус дня'}
              >
                <Star className="w-3.5 h-3.5 fill-current" />
                <span>{isFocused ? 'В фокусе дня' : '⭐ Сегодня'}</span>
              </button>

              <button
                type="button"
                onClick={onClose}
                className="p-1.5 rounded-lg text-muted hover:text-primary hover:bg-canvas transition-colors"
                title="Закрыть (Esc)"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Task Title & Checkbox */}
            <div className="flex items-start gap-3">
              <button
                type="button"
                onClick={() => onToggleTask?.(task.id)}
                className="mt-1 text-muted hover:text-accent-solid transition-colors shrink-0"
              >
                {task.isCompleted ? (
                  <CheckCircle2 className="w-5 h-5 text-accent-solid" />
                ) : (
                  <Circle className="w-5 h-5" />
                )}
              </button>

              <div className="flex-1 min-w-0">
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
                    className="text-lg font-serif bg-canvas/40 border-subtle"
                  />
                ) : (
                  <h2
                    onClick={() => setIsEditingTitle(true)}
                    className={`font-serif text-xl font-normal leading-snug cursor-text hover:bg-canvas/30 rounded p-1 -m-1 transition-colors ${
                      task.isCompleted ? 'line-through text-muted' : 'text-primary'
                    }`}
                    title="Нажмите для редактирования"
                  >
                    {task.title}
                  </h2>
                )}
              </div>
            </div>

            {/* Quick Action Bar */}
            <QuickActionBar
              dueDate={task.dueDate}
              onDueDateChange={(date) =>
                updateTaskMutation.mutate({ taskId: task.id, data: { dueDate: date } })
              }
              onAttachClick={() => fileInputRef.current?.click()}
              onToggleTagInput={() => setIsAddingTag((prev) => !prev)}
              columns={columns}
              selectedColumnId={task.columnId}
              onColumnChange={handleColumnChange}
            />

            {/* Tag form when adding tag */}
            {isAddingTag && (
              <form onSubmit={handleCreateAndAttachTag} className="flex items-center gap-2 pt-1">
                <Input
                  autoFocus
                  placeholder="@контекст..."
                  value={newTagName}
                  onChange={(e) => setNewTagName(e.target.value)}
                  className="text-xs bg-canvas/40 border-subtle h-8"
                />
                <Button type="submit" size="sm" variant="solid" className="h-8 text-xs px-3">
                  Добавить
                </Button>
              </form>
            )}

            {/* Tags row */}
            {task.tags && task.tags.length > 0 && (
              <div className="flex items-center gap-1.5 flex-wrap">
                {task.tags.map((tag: Tag) => (
                  <span
                    key={tag.id}
                    className="inline-flex items-center gap-1 bg-accent-surface text-primary text-[11px] font-medium px-2.5 py-0.5 rounded-full border border-subtle"
                  >
                    #{tag.name}
                    <button
                      type="button"
                      onClick={() => removeTagMutation.mutate({ taskId: task.id, tag })}
                      className="hover:text-muted"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </span>
                ))}
              </div>
            )}

            {/* Priority row */}
            <div className="space-y-1.5">
              <label className="text-[10px] font-mono uppercase tracking-wider text-muted">
                Приоритет
              </label>
              <div className="flex items-center gap-1.5 flex-wrap">
                {PRIORITIES.map((p) => {
                  const active = task.priority === p.value;
                  return (
                    <button
                      key={p.value}
                      type="button"
                      onClick={() =>
                        updateTaskMutation.mutate({ taskId: task.id, data: { priority: p.value } })
                      }
                      className={`text-xs px-2.5 py-1 rounded-full border transition-all ${
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

            {/* Project row */}
            <div className="space-y-1.5">
              <label className="text-[10px] font-mono uppercase tracking-wider text-muted flex items-center gap-1">
                <FolderClosed className="w-3 h-3" /> Проект
              </label>
              <select
                value={task.projectId ?? ''}
                onChange={(e) =>
                  updateTaskMutation.mutate({
                    taskId: task.id,
                    data: { projectId: e.target.value || null },
                  })
                }
                className="w-full text-xs bg-canvas/40 border border-subtle rounded-lg px-2.5 py-1.5 text-primary focus:outline-none focus:border-accent-solid"
              >
                <option value="">(Входящие)</option>
                {projects.map((proj) => (
                  <option key={proj.id} value={proj.id}>
                    {proj.name}
                  </option>
                ))}
              </select>
            </div>

            {/* Markdown Notes with Ctrl+V support */}
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
        ) : null}

        {/* Footer actions */}
        {task && onDeleteTask && (
          <div className="p-4 border-t border-subtle/50 bg-surface flex items-center justify-between">
            <button
              type="button"
              onClick={() => {
                onDeleteTask(task.id);
                onClose();
              }}
              className="flex items-center gap-1.5 text-xs text-muted hover:text-primary transition-colors py-1 px-2 rounded-md hover:bg-canvas"
            >
              <Trash2 className="w-3.5 h-3.5" />
              <span>Удалить задачу</span>
            </button>

            <Button size="sm" variant="outline" onClick={onClose} className="text-xs">
              Готово
            </Button>
          </div>
        )}
      </aside>
    </>
  );
};
