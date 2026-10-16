import React from 'react';
import {
  DndContext,
  DragOverlay,
  closestCorners,
} from '@dnd-kit/core';
import { SortableContext, horizontalListSortingStrategy } from '@dnd-kit/sortable';
import { Skeleton, EmptyState, ErrorBanner } from '@/shared/components/ui';
import { LayoutGrid } from 'lucide-react';
import {
  useProjectBoardsQuery,
  useBoardColumnsQuery,
  useProjectTasksQuery,
  useMoveTaskMutation,
  useReorderColumnMutation,
} from '@/entities/project';
import { useToggleTaskMutation } from '@/features/task-navigation';
import type { Task } from '@/shared/api';
import { KanbanColumn } from './KanbanColumn';
import { KanbanCard } from './KanbanCard';
import { useKanbanDnd } from './useKanbanDnd';

export interface KanbanBoardProps {
  projectId: string;
  activeBoardId?: string | null;
  workspaceId?: string;
  onSelectTask?: (taskId: string) => void;
  selectedTaskId?: string | null;
  onCreateTaskInColumn?: (title: string, columnId: string) => Promise<void>;
  filteredTasks?: Task[];
}

export const KanbanBoard: React.FC<KanbanBoardProps> = ({
  projectId,
  activeBoardId,
  workspaceId,
  onSelectTask,
  selectedTaskId,
  onCreateTaskInColumn,
  filteredTasks,
}) => {
  const boardsQuery = useProjectBoardsQuery(workspaceId, projectId);
  const activeBoard =
    (activeBoardId
      ? boardsQuery.data?.find((b) => b.id === activeBoardId)
      : null) ?? boardsQuery.data?.[0];

  const columnsQuery = useBoardColumnsQuery(workspaceId, activeBoard?.id);
  const columns = [...(columnsQuery.data ?? [])].sort((a, b) => a.position - b.position);

  const tasksQuery = useProjectTasksQuery(workspaceId, projectId);
  const tasks = filteredTasks ?? tasksQuery.data ?? [];

  const moveTaskMutation = useMoveTaskMutation(workspaceId, projectId);
  const reorderColumnMutation = useReorderColumnMutation(workspaceId, activeBoard?.id);
  const toggleTaskMutation = useToggleTaskMutation();

  const handleMoveTask = (taskId: string, columnId: string, position: number) => {
    moveTaskMutation.mutate({
      taskId,
      data: {
        columnId,
        position,
        projectId,
      },
    });
  };

  const handleReorderColumn = (columnId: string, position: number) => {
    reorderColumnMutation.mutate({
      columnId,
      position,
    });
  };

  const {
    activeTask,
    activeColumn,
    sensors,
    handleDragStart,
    handleDragEnd,
    handleDragCancel,
  } = useKanbanDnd({
    tasks,
    columns,
    onMoveTask: handleMoveTask,
    onReorderColumn: handleReorderColumn,
  });

  const isLoading = boardsQuery.isLoading || columnsQuery.isLoading || tasksQuery.isLoading;
  const isError = boardsQuery.isError || columnsQuery.isError || tasksQuery.isError;
  const errorMessage =
    boardsQuery.error?.message ||
    columnsQuery.error?.message ||
    tasksQuery.error?.message ||
    'Не удалось загрузить канбан-доску';

  const handleRetry = () => {
    boardsQuery.refetch();
    columnsQuery.refetch();
    tasksQuery.refetch();
  };

  // State 1: Loading (Skeleton)
  if (isLoading) {
    return (
      <div className="flex gap-4 overflow-x-auto pb-4 h-full">
        {[1, 2, 3].map((idx) => (
          <div
            key={idx}
            className="w-72 shrink-0 bg-canvas/40 border border-subtle/60 rounded-xl p-3 space-y-3"
          >
            <div className="flex items-center justify-between pb-2">
              <Skeleton variant="text" className="w-24 h-4" />
              <Skeleton variant="circle" className="w-5 h-5" />
            </div>
            <Skeleton variant="rect" className="w-full h-20 rounded-lg" />
            <Skeleton variant="rect" className="w-full h-20 rounded-lg" />
            <Skeleton variant="rect" className="w-full h-16 rounded-lg" />
          </div>
        ))}
      </div>
    );
  }

  // State 2: Error
  if (isError) {
    return (
      <div className="py-8">
        <ErrorBanner
          title="Ошибка загрузки доски"
          message={errorMessage}
          onRetry={handleRetry}
        />
      </div>
    );
  }

  // State 3: Empty State (no columns)
  if (columns.length === 0) {
    return (
      <div className="py-12">
        <EmptyState
          icon={<LayoutGrid className="w-7 h-7 stroke-[1.5]" />}
          title="Колонки отсутствуют"
          description="В этом проекте пока нет ни одной колонки для организации задач."
        />
      </div>
    );
  }

  // State 4: Success - Kanban DND Board
  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
    >
      <SortableContext
        items={columns.map((c) => c.id)}
        strategy={horizontalListSortingStrategy}
      >
        <div className="flex gap-4 overflow-x-auto pb-4 h-full min-h-[360px] items-start">
          {columns.map((column) => {
            const columnTasks = tasks
              .filter((t) => t.columnId === column.id)
              .sort((a, b) => a.position - b.position);

            return (
              <KanbanColumn
                key={column.id}
                column={column}
                tasks={columnTasks}
                projectId={projectId}
                workspaceId={workspaceId}
                selectedTaskId={selectedTaskId}
                onSelectTask={onSelectTask}
                onToggleTask={(id) => toggleTaskMutation.mutate(id)}
                onCreateTask={onCreateTaskInColumn}
              />
            );
          })}
        </div>
      </SortableContext>

      <DragOverlay dropAnimation={null}>
        {activeColumn ? (
          <KanbanColumn
            column={activeColumn}
            tasks={tasks
              .filter((t) => t.columnId === activeColumn.id)
              .sort((a, b) => a.position - b.position)}
            projectId={projectId}
            workspaceId={workspaceId}
            isOverlay
          />
        ) : activeTask ? (
          <KanbanCard task={activeTask} isOverlay />
        ) : null}
      </DragOverlay>
    </DndContext>
  );
};
