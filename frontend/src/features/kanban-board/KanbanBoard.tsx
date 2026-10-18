import React, { useState, useRef, useEffect } from 'react';
import {
  DndContext,
  DragOverlay,
  closestCorners,
} from '@dnd-kit/core';
import { SortableContext, horizontalListSortingStrategy } from '@dnd-kit/sortable';
import { Skeleton, EmptyState, ErrorBanner, Button } from '@/shared/components/ui';
import { LayoutGrid, Plus } from 'lucide-react';
import { cn } from '@/shared/lib/cn';
import {
  useProjectBoardsQuery,
  useBoardColumnsQuery,
  useProjectTasksQuery,
  useMoveTaskMutation,
  useReorderColumnMutation,
  useCreateColumnMutation,
} from '@/entities/project';
import { useUpdateTaskMutation } from '@/entities/task';
import { useToggleTaskMutation } from '@/features/task-navigation';
import type { Task } from '@/shared/api';
import { KanbanColumn } from './KanbanColumn';
import { KanbanCard } from './KanbanCard';
import { useKanbanDnd } from './useKanbanDnd';
import { getStoredColumnAutomationConfig } from './ColumnAutomationsModal';

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
  const createColumnMutation = useCreateColumnMutation(workspaceId, activeBoard?.id);
  const updateTaskMutation = useUpdateTaskMutation(workspaceId);
  const toggleTaskMutation = useToggleTaskMutation();

  const [isAddingColumn, setIsAddingColumn] = useState(false);
  const [newColumnName, setNewColumnName] = useState('');
  const [isSubmittingColumn, setIsSubmittingColumn] = useState(false);
  const newColumnInputRef = useRef<HTMLInputElement>(null);

  // Middle-mouse drag pan (D-071-4)
  const scrollContainerRef = useRef<HTMLDivElement | null>(null);
  const [isPanning, setIsPanning] = useState(false);
  const panStartRef = useRef<{ startX: number; scrollLeft: number } | null>(null);

  const handleMouseDown = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.button === 1 && scrollContainerRef.current) {
      e.preventDefault();
      setIsPanning(true);
      panStartRef.current = {
        startX: e.clientX,
        scrollLeft: scrollContainerRef.current.scrollLeft,
      };
    }
  };

  useEffect(() => {
    if (!isPanning) return;

    const handleMouseMove = (e: MouseEvent) => {
      if (!panStartRef.current || !scrollContainerRef.current) return;
      const dx = e.clientX - panStartRef.current.startX;
      scrollContainerRef.current.scrollLeft = panStartRef.current.scrollLeft - dx;
    };

    const handleMouseUp = (e: MouseEvent) => {
      if (e.button === 1 || e.buttons === 0) {
        setIsPanning(false);
        panStartRef.current = null;
      }
    };

    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseup', handleMouseUp);
    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isPanning]);

  const handleCreateColumn = async () => {
    const trimmed = newColumnName.trim();
    if (!trimmed || isSubmittingColumn) return;

    setIsSubmittingColumn(true);
    try {
      await createColumnMutation.mutateAsync({ name: trimmed });
      setNewColumnName('');
      setIsAddingColumn(false);
    } finally {
      setIsSubmittingColumn(false);
    }
  };

  const handleMoveTask = (taskId: string, columnId: string, position: number) => {
    moveTaskMutation.mutate({
      taskId,
      data: {
        columnId,
        position,
        projectId,
      },
    });

    // Column automation: change priority upon entering column
    const autoConfig = getStoredColumnAutomationConfig(columnId);
    if (autoConfig.enterPriorityEnabled && autoConfig.enterPriority) {
      updateTaskMutation.mutate({
        taskId,
        data: { priority: autoConfig.enterPriority },
      });
    }
  };

  const handleToggleTask = (taskId: string) => {
    const task = tasks.find((t) => t.id === taskId);
    toggleTaskMutation.mutate(taskId);

    // If task was incomplete, and is now being completed:
    if (task && !task.isCompleted && task.columnId) {
      const autoConfig = getStoredColumnAutomationConfig(task.columnId);
      if (autoConfig.completeMoveEnabled && autoConfig.completeTargetColumnId) {
        // Automatically move task to target column upon completion!
        moveTaskMutation.mutate({
          taskId,
          data: {
            columnId: autoConfig.completeTargetColumnId,
            position: Date.now(),
            projectId,
          },
        });
      }
    }
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
  if (columns.length === 0 && !isAddingColumn) {
    return (
      <div className="py-12 flex flex-col items-center justify-center">
        <EmptyState
          icon={<LayoutGrid className="w-7 h-7 stroke-[1.5]" />}
          title="Колонки отсутствуют"
          description="В этом проекте пока нет ни одной колонки для организации задач."
        />
        <div className="mt-4">
          <Button
            variant="solid"
            size="sm"
            onClick={() => {
              setIsAddingColumn(true);
              setTimeout(() => newColumnInputRef.current?.focus(), 50);
            }}
          >
            <Plus className="w-4 h-4 mr-1.5" />
            Добавить колонку
          </Button>
        </div>
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
        <div
          ref={scrollContainerRef}
          onMouseDown={handleMouseDown}
          className={cn(
            'flex gap-4 overflow-x-auto pb-4 h-full min-h-[360px] items-start',
            isPanning && 'cursor-grabbing select-none'
          )}
        >
          {columns.map((column) => {
            const columnTasks = tasks
              .filter((t) => t.columnId === column.id)
              .sort((a, b) => a.position - b.position);

            return (
              <KanbanColumn
                key={column.id}
                column={column}
                tasks={columnTasks}
                allColumns={columns}
                projectId={projectId}
                workspaceId={workspaceId}
                selectedTaskId={selectedTaskId}
                onSelectTask={onSelectTask}
                onToggleTask={handleToggleTask}
                onCreateTask={onCreateTaskInColumn}
              />
            );
          })}

          {/* Ghost Button: Добавить колонку (равная ширина w-72) */}
          {!isAddingColumn ? (
            <button
              type="button"
              onClick={() => {
                setIsAddingColumn(true);
                setTimeout(() => newColumnInputRef.current?.focus(), 50);
              }}
              className="w-72 shrink-0 min-h-[140px] rounded-xl border-2 border-dashed border-subtle/70 hover:border-accent-solid/60 bg-transparent hover:bg-canvas/50 flex flex-col items-center justify-center gap-2 text-muted hover:text-primary transition-all duration-150 group cursor-pointer select-none"
              title="Добавить новую колонку на доску"
            >
              <div className="w-8 h-8 rounded-lg bg-surface border border-subtle/60 flex items-center justify-center group-hover:border-accent-solid/50 group-hover:text-accent-solid transition-colors">
                <Plus className="w-4 h-4" />
              </div>
              <span className="text-xs font-medium">Добавить колонку</span>
            </button>
          ) : (
            <div className="w-72 shrink-0 bg-canvas/80 border border-subtle rounded-xl p-3 shadow-washi space-y-2.5 animate-in fade-in duration-150">
              <div className="text-xs font-semibold text-primary uppercase tracking-wider">
                Новая колонка
              </div>
              <input
                ref={newColumnInputRef}
                type="text"
                value={newColumnName}
                onChange={(e) => setNewColumnName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleCreateColumn();
                  if (e.key === 'Escape') {
                    setIsAddingColumn(false);
                    setNewColumnName('');
                  }
                }}
                placeholder="Название колонки..."
                className="w-full text-xs px-2.5 py-1.5 rounded-lg border border-subtle bg-surface text-primary placeholder:text-muted/60 focus:outline-none focus:ring-1 focus:ring-accent-solid"
                autoFocus
              />
              <div className="flex items-center justify-end gap-1.5 pt-1">
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-xs h-7 px-2"
                  onClick={() => {
                    setIsAddingColumn(false);
                    setNewColumnName('');
                  }}
                >
                  Отмена
                </Button>
                <Button
                  variant="solid"
                  size="sm"
                  className="text-xs h-7 px-3"
                  disabled={!newColumnName.trim() || isSubmittingColumn}
                  onClick={handleCreateColumn}
                >
                  Создать
                </Button>
              </div>
            </div>
          )}
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
