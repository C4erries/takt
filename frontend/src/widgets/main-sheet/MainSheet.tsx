import React from 'react';
import type { Task, CreateInboxTaskRequest } from '@/shared/api';
import { useTaskStore } from '@/features/task-navigation';
import {
  useInboxTasks,
  useCreateTaskMutation,
  useToggleTaskMutation,
  useDeleteTaskMutation,
  useTodayTasksQuery,
} from '@/entities/task';
import { InboxView } from './InboxView';
import { TodayView } from './TodayView';
import { ProjectView } from './ProjectView';
import { AutomationsHub } from '@/features/automations';
import { ArchiveView } from '@/features/archive';
import { TagManagerView } from '@/features/tag-manager';

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
  const todayQuery = useTodayTasksQuery(workspaceId);
  const createTaskMutation = useCreateTaskMutation();
  const toggleTaskMutation = useToggleTaskMutation();
  const deleteTaskMutation = useDeleteTaskMutation();
  const {
    activeNav,
    selectedProjectId,
    selectedTaskId,
    activeBoardId,
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
      const targetList = activeNav === 'today' ? (todayQuery.data ?? []) : currentTasks;
      if (targetList[index]) {
        setSelectedTaskId(targetList[index].id);
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
    <div className="flex-1 w-full bg-surface border border-subtle rounded-sheet shadow-washi flex flex-col overflow-hidden my-3">
      {isProjectNav && activeProjectId ? (
        <ProjectView
          projectId={activeProjectId}
          activeBoardId={activeBoardId}
          workspaceId={workspaceId}
          selectedTaskId={selectedTaskId}
          onSelectTask={(id) => setSelectedTaskId(id)}
        />
      ) : activeNav === 'today' ? (
        <TodayView
          tasks={todayQuery.data ?? []}
          isLoading={todayQuery.isLoading}
          error={todayQuery.error ? todayQuery.error.message : null}
          onRetry={() => todayQuery.refetch()}
          selectedIndex={currentSelectedIndex}
          onSelectTask={handleSelectTask}
          onToggleTask={handleToggleTask}
          onDeleteTask={handleDeleteTask}
          onCreateTodayTask={handleCreateTask}
          onOpenQuickCapture={handleOpenQuickCapture}
        />
      ) : activeNav === 'automations' ? (
        <AutomationsHub workspaceId={workspaceId} />
      ) : activeNav === 'archive' ? (
        <ArchiveView workspaceId={workspaceId} />
      ) : activeNav === 'tags' ? (
        <TagManagerView workspaceId={workspaceId} />
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
