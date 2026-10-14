import React from 'react';
import type { Task, CreateInboxTaskRequest } from '@/shared/api';
import {
  useInboxTasks,
  useCreateTaskMutation,
  useToggleTaskMutation,
  useDeleteTaskMutation,
  useTaskStore,
} from '@/features/task-navigation';
import { InboxView } from './InboxView';
import { ProjectView } from './ProjectView';

export interface MainSheetProps {
  tasks?: Task[];
  isLoading?: boolean;
  error?: string | null;
  onRetry?: () => void;
  selectedIndex?: number;
  onSelectTask?: (index: number) => void;
  onToggleTask?: (taskId: string) => void;
  onDeleteTask?: (taskId: string) => void;
  onCreateTask?: (req: CreateInboxTaskRequest) => Promise<void>;
  onOpenQuickCapture?: () => void;
  workspaceId?: string;
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
  workspaceId,
}) => {
  const query = useInboxTasks();
  const createTaskMutation = useCreateTaskMutation();
  const toggleTaskMutation = useToggleTaskMutation();
  const deleteTaskMutation = useDeleteTaskMutation();
  const {
    activeNav,
    selectedProjectId,
    selectedTaskId,
    setSelectedTaskId,
    setQuickCaptureOpen,
  } = useTaskStore();

  const currentTasks = tasks ?? query.data ?? [];
  const currentLoading = isLoading !== undefined ? isLoading : query.isLoading;
  const currentError = error !== undefined ? error : (query.error ? query.error.message : null);
  const handleRetry = onRetry ?? (() => query.refetch());

  const currentSelectedIndex =
    selectedIndex !== undefined
      ? selectedIndex
      : Math.max(0, currentTasks.findIndex((t) => t.id === selectedTaskId));

  const handleSelectTask =
    onSelectTask ??
    ((index: number) => {
      if (currentTasks[index]) {
        setSelectedTaskId(currentTasks[index].id);
      }
    });

  const handleToggleTask =
    onToggleTask ?? ((taskId: string) => toggleTaskMutation.mutate(taskId));

  const handleDeleteTask =
    onDeleteTask ?? ((taskId: string) => deleteTaskMutation.mutate(taskId));

  const handleCreateTask =
    onCreateTask ??
    (async (req: CreateInboxTaskRequest) => {
      await createTaskMutation.mutateAsync(req);
    });

  const handleOpenQuickCapture =
    onOpenQuickCapture ?? (() => setQuickCaptureOpen(true));

  // Determine if a project is active
  const isProjectNav = activeNav.startsWith('project:') || Boolean(selectedProjectId);
  const activeProjectId =
    selectedProjectId ||
    (activeNav.startsWith('project:') ? activeNav.replace('project:', '') : null);

  return (
    <div className="flex-1 max-w-[880px] w-full bg-surface border border-subtle rounded-sheet shadow-washi flex flex-col overflow-hidden my-3">
      {isProjectNav && activeProjectId ? (
        <ProjectView
          projectId={activeProjectId}
          workspaceId={workspaceId}
          selectedTaskId={selectedTaskId}
          onSelectTask={(id) => setSelectedTaskId(id)}
        />
      ) : (
        <InboxView
          tasks={currentTasks}
          isLoading={currentLoading}
          error={currentError}
          onRetry={handleRetry}
          selectedIndex={currentSelectedIndex}
          onSelectTask={handleSelectTask}
          onToggleTask={handleToggleTask}
          onDeleteTask={handleDeleteTask}
          onCreateTask={handleCreateTask}
          onOpenQuickCapture={handleOpenQuickCapture}
        />
      )}
    </div>
  );
};
