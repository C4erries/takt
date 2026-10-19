import React, { useState, useEffect, useRef, useMemo } from 'react';
import {
  X,
  Star,
  CheckCircle2,
  Circle,
  Trash2,
  FolderClosed,
} from 'lucide-react';
import type { Priority, Task } from '@/shared/api';
import {
  useUpdateTaskMutation,
  useSetTaskFocusMutation,
} from '@/entities/task';
import {
  useProjectsQuery,
  useProjectBoardsQuery,
  useBoardColumnsQuery,
  useProjectAllColumnsQuery,
  useMoveTaskMutation,
} from '@/entities/project';
import { useUploadAttachmentMutation } from '@/entities/attachment';
import { MarkdownNotes, TaskAttachments } from '@/widgets/task-inspector';
import { Button, Input } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import { QuickActionBar } from './QuickActionBar';
import { TaskTagSection } from './TaskTagSection';
import { TaskAutomationsSection } from './TaskAutomationsSection';
import type { TaskDrawerProps } from './types';

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

export const TaskDrawer: React.FC<TaskDrawerProps> = ({
  task,
  workspaceId,
  onClose,
  onToggleTask,
  onDeleteTask,
}) => {
  const isOpen = Boolean(task);

  const lastTaskRef = useRef<Task | null>(task);
  if (task) {
    lastTaskRef.current = task;
  }
  const currentTask = task || lastTaskRef.current;

  const updateTaskMutation = useUpdateTaskMutation(workspaceId);
  const setFocusMutation = useSetTaskFocusMutation(workspaceId);
  const { data: projects = [] } = useProjectsQuery(workspaceId);

  const boardsQuery = useProjectBoardsQuery(workspaceId, currentTask?.projectId ?? undefined);
  const boards = boardsQuery.data ?? [];
  const activeBoard = boards[0];
  const { data: boardGroups = [] } = useProjectAllColumnsQuery(workspaceId, boards);
  const columnsQuery = useBoardColumnsQuery(workspaceId, activeBoard?.id);
  const columns = columnsQuery.data ?? [];
  const allColumns = useMemo(
    () => (boardGroups.length > 0 ? boardGroups.flatMap((bg) => bg.columns) : columns),
    [boardGroups, columns]
  );
  const moveTaskMutation = useMoveTaskMutation(workspaceId, currentTask?.projectId ?? undefined);

  const uploadAttachmentMutation = useUploadAttachmentMutation(workspaceId, currentTask?.id);

  const [title, setTitle] = useState(currentTask?.title ?? '');
  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [isAddingTag, setIsAddingTag] = useState(false);

  const fileInputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    if (task) {
      setTitle(task.title ?? '');
      setIsEditingTitle(false);
      setIsAddingTag(false);
    }
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

  if (!currentTask) return null;

  const isFocused = currentTask.focusOrder !== null && currentTask.focusOrder !== undefined;

  const handleTitleBlur = () => {
    setIsEditingTitle(false);
    if (!currentTask) return;
    const trimmed = title.trim();
    if (trimmed && trimmed !== currentTask.title) {
      updateTaskMutation.mutate({ taskId: currentTask.id, data: { title: trimmed } });
    }
  };

  const handleUploadImageForNotes = async (file: File): Promise<string | null> => {
    if (!currentTask) return null;
    try {
      const att = await uploadAttachmentMutation.mutateAsync(file);
      return att.contentUrl;
    } catch {
      return null;
    }
  };

  const handleFileInputChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0 && currentTask) {
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
    if (!currentTask || !columnId) return;
    moveTaskMutation.mutate({
      taskId: currentTask.id,
      data: {
        columnId,
        position: currentTask.position,
        projectId: currentTask.projectId ?? undefined,
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

        {currentTask ? (
          <div className="flex-1 flex flex-col overflow-y-auto p-6 space-y-5">
            {/* Header top bar */}
            <div className="flex items-center justify-between border-b border-subtle/50 pb-3">
              <button
                type="button"
                onClick={() =>
                  setFocusMutation.mutate({
                    taskId: currentTask.id,
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
                onClick={() => onToggleTask?.(currentTask.id)}
                className="mt-1 text-muted hover:text-accent-solid transition-colors shrink-0"
              >
                {currentTask.isCompleted ? (
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
                        setTitle(currentTask.title);
                        setIsEditingTitle(false);
                      }
                    }}
                    className="text-lg font-serif bg-canvas/40 border-subtle"
                  />
                ) : (
                  <h2
                    onClick={() => setIsEditingTitle(true)}
                    className={`font-serif text-xl font-normal leading-snug cursor-text hover:bg-canvas/30 rounded p-1 -m-1 transition-colors ${
                      currentTask.isCompleted ? 'line-through text-muted' : 'text-primary'
                    }`}
                    title="Нажмите для редактирования"
                  >
                    {currentTask.title}
                  </h2>
                )}
              </div>
            </div>

            {/* Quick Action Bar */}
            <QuickActionBar
              dueDate={currentTask.dueDate}
              onDueDateChange={(date) =>
                updateTaskMutation.mutate({ taskId: currentTask.id, data: { dueDate: date } })
              }
              onAttachClick={() => fileInputRef.current?.click()}
              onToggleTagInput={() => setIsAddingTag((prev) => !prev)}
              columns={columns}
              boardGroups={boardGroups}
              selectedColumnId={currentTask.columnId}
              onColumnChange={handleColumnChange}
            />

            {/* Tags section with autocomplete and color palette */}
            <TaskTagSection
              taskId={currentTask.id}
              tags={currentTask.tags}
              workspaceId={workspaceId}
              isAddingTag={isAddingTag}
              onToggleAddingTag={() => setIsAddingTag((prev) => !prev)}
            />

            {/* Priority row with self-colored active selection (D-071-3) */}
            <div className="space-y-1.5">
              <label className="text-[10px] font-mono uppercase tracking-wider text-muted">
                Приоритет
              </label>
              <div className="flex items-center gap-1.5 flex-wrap">
                {PRIORITIES.map((p) => {
                  const active =
                    currentTask.priority === p.value ||
                    (p.value === 'pending' && currentTask.priority === 'urgent');
                  return (
                    <button
                      key={p.value}
                      type="button"
                      onClick={() =>
                        updateTaskMutation.mutate({ taskId: currentTask.id, data: { priority: p.value } })
                      }
                      className={cn(
                        'text-xs px-2.5 py-1 rounded-full border transition-all',
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

            {/* Project row */}
            <div className="space-y-1.5">
              <label className="text-[10px] font-mono uppercase tracking-wider text-muted flex items-center gap-1">
                <FolderClosed className="w-3 h-3" /> Проект
              </label>
              <select
                value={currentTask.projectId ?? ''}
                onChange={(e) =>
                  updateTaskMutation.mutate({
                    taskId: currentTask.id,
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
              initialValue={currentTask.descriptionMarkdown ?? ''}
              onSave={(val) =>
                updateTaskMutation.mutate({ taskId: currentTask.id, data: { descriptionMarkdown: val } })
              }
              onUploadImage={handleUploadImageForNotes}
            />

            {/* Attachments Section */}
            <TaskAttachments workspaceId={workspaceId} taskId={currentTask.id} />

            {/* Task automations (D-072-4) */}
            <TaskAutomationsSection
              taskId={currentTask.id}
              workspaceId={workspaceId}
              columnId={currentTask.columnId}
              boardId={activeBoard?.id}
              projectId={currentTask.projectId}
              allColumns={allColumns}
              boardColumnGroups={boardGroups}
            />
          </div>
        ) : null}

        {/* Footer actions */}
        {currentTask && onDeleteTask && (
          <div className="p-4 border-t border-subtle/50 bg-surface flex items-center justify-between">
            <button
              type="button"
              onClick={() => {
                onDeleteTask(currentTask.id);
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
