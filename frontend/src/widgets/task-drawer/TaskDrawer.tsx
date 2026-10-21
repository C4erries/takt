import React, { useState, useEffect, useRef, useMemo } from 'react';
import { X, Star, CheckCircle2, Circle, Trash2, FolderClosed, Clock, CornerUpLeft } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import type { Task } from '@/shared/api';
import {
  useUpdateTaskMutation,
  useSetTaskFocusMutation,
  useSubtasksQuery,
  formatTaskNumber,
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
import { motion, AnimatePresence, drawerVariants, backdropVariants } from '@/shared/lib/motion';
import { QuickActionBar } from './QuickActionBar';
import { TaskTagSection } from './TaskTagSection';
import { TaskSubtasksSection } from './TaskSubtasksSection';
import { TaskAutomationsSection } from './TaskAutomationsSection';
import { type TaskDrawerProps, PRIORITIES, PRIORITY_ACTIVE_STYLES } from './types';

export const TaskDrawer: React.FC<TaskDrawerProps> = ({
  task,
  workspaceId,
  onClose,
  onToggleTask,
  onDeleteTask,
  onSelectTask,
}) => {
  const isOpen = Boolean(task);
  const lastTaskRef = useRef<Task | null>(task);
  if (task) lastTaskRef.current = task;
  const currentTask = task || lastTaskRef.current;

  const queryClient = useQueryClient();
  const updateTaskMutation = useUpdateTaskMutation(workspaceId);
  const setFocusMutation = useSetTaskFocusMutation(workspaceId);
  const { data: projects = [] } = useProjectsQuery(workspaceId);

  const { data: subtasks = [] } = useSubtasksQuery(workspaceId, currentTask?.id);

  const parentTask = useMemo(() => {
    if (!currentTask?.parentId) return null;
    const allQueries = queryClient.getQueriesData<Task[]>({ queryKey: [] });
    for (const [, tasksList] of allQueries) {
      if (Array.isArray(tasksList)) {
        const found = tasksList.find((t) => t && t.id === currentTask.parentId);
        if (found) return found;
      }
    }
    return null;
  }, [currentTask?.parentId, queryClient]);

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
  const drawerRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (task) {
      setTitle(task.title ?? '');
      setIsEditingTitle(false);
      setIsAddingTag(false);
    }
  }, [task?.id, task?.title]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen && !isEditingTitle && !isAddingTag) onClose();
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, isEditingTitle, isAddingTag, onClose]);

  useEffect(() => {
    if (!isOpen) return;
    const handleClickOutside = (e: MouseEvent) => {
      if (drawerRef.current && !drawerRef.current.contains(e.target as Node)) onClose();
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen, onClose]);

  const isFocused = currentTask?.focusOrder !== null && currentTask?.focusOrder !== undefined;

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
        try { await uploadAttachmentMutation.mutateAsync(file); } catch {}
      }
      e.target.value = '';
    }
  };

  const handleColumnChange = (columnId: string) => {
    if (!currentTask || !columnId) return;
    moveTaskMutation.mutate({
      taskId: currentTask.id,
      data: { columnId, position: currentTask.position, projectId: currentTask.projectId ?? undefined },
    });
  };

  return (
    <AnimatePresence>
      {isOpen && currentTask && (
        <motion.div
          key="drawer-backdrop"
          variants={backdropVariants}
          initial="hidden"
          animate="visible"
          exit="exit"
          className="fixed inset-0 bg-primary/20 z-40 lg:hidden"
          onClick={onClose}
        />
      )}

      {isOpen && currentTask && (
        <motion.aside
          ref={drawerRef}
          key="drawer-panel"
          variants={drawerVariants}
          initial="hidden"
          animate="visible"
          exit="exit"
          className="fixed top-0 right-0 h-full w-full max-w-[560px] bg-surface z-40 border-l border-subtle shadow-washi-modal flex flex-col justify-between"
        >
            <input type="file" ref={fileInputRef} onChange={handleFileInputChange} className="hidden" multiple />

            <div className="flex-1 flex flex-col overflow-y-auto p-6 space-y-5">
              {/* Header top bar */}
              <div className="flex items-center justify-between border-b border-subtle/50 pb-3">
                <div className="flex items-center gap-2 flex-wrap">
                  <button
                    type="button"
                    onClick={() => setFocusMutation.mutate({ taskId: currentTask.id, focusOrder: isFocused ? null : 1 })}
                    className={cn(
                      'flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium transition-colors',
                      isFocused ? 'bg-accent-solid text-surface' : 'bg-canvas text-muted hover:text-primary hover:bg-canvas/80'
                    )}
                    title={isFocused ? 'Снять из фокуса дня' : 'Добавить в фокус дня'}
                  >
                    <Star className="w-3.5 h-3.5 fill-current" />
                    <span>{isFocused ? 'В фокусе дня' : '⭐ Сегодня'}</span>
                  </button>

                  {/* WEEEK-style breadcrumbs #parent > #child or task number badge */}
                  {currentTask.parentId ? (
                    <div className="flex items-center gap-1 font-mono text-xs text-muted bg-canvas/60 px-2 py-0.5 rounded border border-subtle/60">
                      <button
                        type="button"
                        onClick={() => onSelectTask?.(currentTask.parentId!)}
                        className="hover:text-accent-solid transition-colors"
                        title={parentTask?.title ? `Перейти к: ${parentTask.title}` : 'Родительская задача'}
                      >
                        {formatTaskNumber(currentTask.parentId)}
                      </button>
                      <span className="text-muted/50">&gt;</span>
                      <span className="text-primary font-medium">{formatTaskNumber(currentTask.id)}</span>
                    </div>
                  ) : (
                    <span className="font-mono text-xs text-muted bg-canvas/40 px-2 py-0.5 rounded border border-subtle/40">
                      {formatTaskNumber(currentTask.id)}
                    </span>
                  )}
                </div>

                <button
                  type="button"
                  onClick={onClose}
                  className="p-1.5 rounded-lg text-muted hover:text-primary hover:bg-canvas transition-colors"
                  title="Закрыть (Esc)"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>

              {/* Task Title, Checkbox & Quick Indicator */}
              <div className="space-y-1.5">
                <div className="flex items-start gap-3">
                  <button
                    type="button"
                    onClick={() => onToggleTask?.(currentTask.id)}
                    className="mt-1 text-muted hover:text-accent-solid transition-colors shrink-0"
                  >
                    {currentTask.isCompleted ? <CheckCircle2 className="w-5 h-5 text-accent-solid" /> : <Circle className="w-5 h-5" />}
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
                          if (e.key === 'Escape') { setTitle(currentTask.title); setIsEditingTitle(false); }
                        }}
                        className="text-lg font-serif bg-surface border-subtle"
                      />
                    ) : (
                      <h2
                        onClick={() => setIsEditingTitle(true)}
                        className={cn(
                          'font-serif text-xl font-normal leading-snug cursor-text hover:bg-canvas/30 rounded p-1 -m-1 transition-colors',
                          currentTask.isCompleted ? 'line-through text-muted' : 'text-primary'
                        )}
                        title="Нажмите для редактирования"
                      >
                        {currentTask.title}
                      </h2>
                    )}
                  </div>
                </div>

                {/* Quick subtasks indicator ⏱ X/Y */}
                {(subtasks.length > 0 || (currentTask.subtasksCount ?? 0) > 0) && (
                  <div className="flex items-center gap-1.5 text-xs text-muted font-mono pl-8">
                    <Clock className="w-3.5 h-3.5 text-muted shrink-0" />
                    <span>
                      {subtasks.length > 0
                        ? `${subtasks.filter((s) => s.isCompleted).length}/${subtasks.length}`
                        : `${currentTask.subtasksCompletedCount ?? 0}/${currentTask.subtasksCount ?? 0}`}
                    </span>
                  </div>
                )}
              </div>

              {/* Quick Action Bar */}
              <QuickActionBar
                dueDate={currentTask.dueDate}
                onDueDateChange={(date) => updateTaskMutation.mutate({ taskId: currentTask.id, data: { dueDate: date } })}
                onAttachClick={() => fileInputRef.current?.click()}
                onToggleTagInput={() => setIsAddingTag((prev) => !prev)}
                columns={columns}
                boardGroups={boardGroups}
                selectedColumnId={currentTask.columnId}
                onColumnChange={handleColumnChange}
              />

              {/* Back button to parent task if subtask */}
              {currentTask.parentId && (
                <div>
                  <button
                    type="button"
                    onClick={() => onSelectTask?.(currentTask.parentId!)}
                    className="inline-flex items-center gap-1.5 text-xs font-medium text-accent-solid hover:text-accent-solid-hover bg-accent-surface/30 hover:bg-accent-surface/50 border border-accent-solid/30 rounded-lg px-2.5 py-1.5 transition-colors max-w-full truncate"
                    title={parentTask?.title ? `Вернуться к: ${parentTask.title}` : 'Родительская задача'}
                  >
                    <CornerUpLeft className="w-3.5 h-3.5 shrink-0" />
                    <span className="truncate">{parentTask?.title || 'Родительская задача'}</span>
                  </button>
                </div>
              )}

              {/* Subtasks Section */}
              <TaskSubtasksSection
                workspaceId={workspaceId}
                parentTaskId={currentTask.id}
                onOpenTask={onSelectTask}
                onToggleTask={onToggleTask}
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
                <label className="text-[10px] font-mono uppercase tracking-wider text-muted">Приоритет</label>
                <div className="flex items-center gap-1.5 flex-wrap">
                  {PRIORITIES.map((p) => {
                    const active = currentTask.priority === p.value || (p.value === 'pending' && currentTask.priority === 'urgent');
                    return (
                      <button
                        key={p.value}
                        type="button"
                        onClick={() => updateTaskMutation.mutate({ taskId: currentTask.id, data: { priority: p.value } })}
                        className={cn('text-xs px-2.5 py-1 rounded-full border transition-all', active ? `${PRIORITY_ACTIVE_STYLES[p.value]} font-medium shadow-washi-sm` : 'border-subtle/60 text-muted hover:text-primary hover:bg-canvas/50')}
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
                  onChange={(e) => updateTaskMutation.mutate({ taskId: currentTask.id, data: { projectId: e.target.value || null } })}
                  className="w-full text-xs bg-surface border border-subtle rounded-lg px-2.5 py-1.5 text-primary focus:outline-none focus:border-accent-solid"
                >
                  <option value="">(Входящие)</option>
                  {projects.map((proj) => <option key={proj.id} value={proj.id}>{proj.name}</option>)}
                </select>
              </div>

              {/* Markdown Notes with Ctrl+V support */}
              <MarkdownNotes
                initialValue={currentTask.descriptionMarkdown ?? ''}
                onSave={(val) => updateTaskMutation.mutate({ taskId: currentTask.id, data: { descriptionMarkdown: val } })}
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

            {/* Footer actions */}
            {onDeleteTask && (
              <div className="p-4 border-t border-subtle/50 bg-surface flex items-center justify-between">
                <button
                  type="button"
                  onClick={() => { onDeleteTask(currentTask.id); onClose(); }}
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
          </motion.aside>
        )}
    </AnimatePresence>
  );
};
