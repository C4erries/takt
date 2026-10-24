import React, { useState, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { FolderClosed, Plus, Filter } from 'lucide-react';
import { ViewToggle, KanbanBoard, getStoredColumnAutomationConfig } from '@/features/kanban-board';
import { useTaskStore } from '@/features/task-navigation';
import {
  useProjectsQuery,
  useProjectBoardsQuery,
  useBoardColumnsQuery,
  useProjectTasksQuery,
  projectApi,
  DEMO_PROJECT_TASKS,
} from '@/entities/project';
import { useToggleTaskMutation, useDeleteTaskMutation } from '@/entities/task';
import { BoardFilterBar, filterAndSortTasks, initialFilterState } from '@/features/board-filters';
import type { BoardFilterState } from '@/features/board-filters';
import { getAuthToken, tasksApi, type Task } from '@/shared/api';
import { TaskRow } from './TaskRow';
import { Skeleton, EmptyState, ErrorBanner, Input, Button } from '@/shared/components/ui';

export interface ProjectViewProps {
  projectId: string;
  activeBoardId?: string | null;
  workspaceId?: string;
  onSelectTask?: (taskId: string) => void;
  selectedTaskId?: string | null;
}

export const ProjectView: React.FC<ProjectViewProps> = ({
  projectId,
  activeBoardId,
  workspaceId,
  onSelectTask,
  selectedTaskId,
}) => {
  const queryClient = useQueryClient();
  const { viewMode, setViewMode } = useTaskStore();
  const { data: projects = [] } = useProjectsQuery(workspaceId);
  const currentProject = projects.find((p) => p.id === projectId) ?? projects[0];

  const boardsQuery = useProjectBoardsQuery(workspaceId, projectId);
  const activeBoard =
    (activeBoardId
      ? boardsQuery.data?.find((b) => b.id === activeBoardId)
      : null) ?? boardsQuery.data?.[0];

  const tasksQuery = useProjectTasksQuery(workspaceId, projectId);
  const tasks = tasksQuery.data ?? [];

  const [filters, setFilters] = useState<BoardFilterState>(initialFilterState);
  const visibleTasks = useMemo(() => filterAndSortTasks(tasks, filters), [tasks, filters]);

  const toggleTaskMutation = useToggleTaskMutation();
  const deleteTaskMutation = useDeleteTaskMutation();

  const [inlineTitle, setInlineTitle] = useState('');

  const columnsQuery = useBoardColumnsQuery(workspaceId, activeBoard?.id);
  const firstColumn = columnsQuery.data?.[0];

  const handleCreateTaskInColumn = async (title: string, columnId: string, description?: string) => {
    const autoConfig = getStoredColumnAutomationConfig(columnId);
    const initialPriority = autoConfig.enterPriorityEnabled ? autoConfig.enterPriority : 'none';

    if (!workspaceId || !getAuthToken()) {
      const newTask: Task = {
        id: `pt-${Date.now()}`,
        workspaceId: workspaceId || 'demo-workspace',
        projectId,
        columnId,
        title,
        descriptionMarkdown: description || '',
        isCompleted: false,
        priority: initialPriority,
        position: Date.now(),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      DEMO_PROJECT_TASKS.push(newTask);
      queryClient.setQueryData<Task[]>(
        ['project-tasks', workspaceId, projectId],
        (old = []) => [...old, newTask]
      );
      return;
    }

    try {
      const created = await tasksApi.createInbox({
        title,
        descriptionMarkdown: description,
        priority: initialPriority,
      });
      if (created?.id) {
        await projectApi.moveTask(workspaceId, created.id, {
          columnId,
          position: Date.now(),
          projectId,
        });
        queryClient.invalidateQueries({ queryKey: ['project-tasks', workspaceId, projectId] });
      }
    } catch {
      // Invalidation or error handled
    }
  };

  const handleInlineSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = inlineTitle.trim();
    if (!trimmed) return;

    if (!workspaceId || !getAuthToken()) {
      const newTask: Task = {
        id: `pt-${Date.now()}`,
        workspaceId: workspaceId || 'demo-workspace',
        projectId,
        columnId: firstColumn?.id,
        title: trimmed,
        descriptionMarkdown: '',
        isCompleted: false,
        priority: 'none',
        position: Date.now(),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      DEMO_PROJECT_TASKS.push(newTask);
      queryClient.setQueryData<Task[]>(
        ['project-tasks', workspaceId, projectId],
        (old = []) => [...old, newTask]
      );
      setInlineTitle('');
      return;
    }

    try {
      const created = await tasksApi.createInbox({ title: trimmed });
      if (created?.id && firstColumn?.id) {
        await projectApi.moveTask(workspaceId, created.id, {
          columnId: firstColumn.id,
          position: Date.now(),
          projectId,
        });
        queryClient.invalidateQueries({ queryKey: ['project-tasks', workspaceId, projectId] });
      }
      setInlineTitle('');
    } catch {
      // Handled
    }
  };

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* 1. Compact Navbar (44px / h-11) */}
      <header className="h-11 shrink-0 px-6 border-b border-subtle/60 flex items-center justify-between bg-surface/90 select-none">
        {/* Breadcrumbs: [Project Name] / [Board Name] */}
        <div className="flex items-center gap-2 min-w-0 text-xs">
          <span className="w-2 h-2 rounded-full bg-accent-solid inline-block shrink-0" />
          <span className="font-semibold text-primary truncate max-w-[200px]">
            {currentProject?.name || 'Проект'}
          </span>
          <span className="text-muted/50 font-normal">/</span>
          <span className="text-muted font-medium truncate max-w-[180px]">
            {activeBoard?.name || 'Основная'}
          </span>
        </div>

        {/* View Toggle */}
        <div className="flex items-center gap-2">
          <ViewToggle view={viewMode} onViewChange={setViewMode} />
        </div>
      </header>

      {/* 2. Dynamic Filters Row */}
      <div className="px-6 py-1.5 border-b border-subtle/40 bg-surface/40">
        <BoardFilterBar
          filters={filters}
          onFilterChange={setFilters}
          workspaceId={workspaceId}
        />
      </div>

      {/* 3. Project Content Area */}
      <div className="flex-1 overflow-y-auto p-6">
        {viewMode === 'kanban' ? (
          <KanbanBoard
            projectId={projectId}
            activeBoardId={activeBoard?.id}
            workspaceId={workspaceId}
            selectedTaskId={selectedTaskId}
            onSelectTask={onSelectTask}
            onCreateTaskInColumn={handleCreateTaskInColumn}
            filteredTasks={visibleTasks}
          />
        ) : (
          <div className="space-y-4 max-w-3xl">
            {/* Inline Quick Capture row for List view */}
            <form onSubmit={handleInlineSubmit} className="flex items-center gap-2">
              <Input
                value={inlineTitle}
                onChange={(e) => setInlineTitle(e.target.value)}
                placeholder="Новая задача проекта... (Enter)"
                className="bg-canvas/50 border-subtle focus-within:bg-surface text-sm"
              />
              <Button
                type="submit"
                variant="solid"
                size="md"
                disabled={!inlineTitle.trim()}
                leftIcon={<Plus className="w-4 h-4" />}
              >
                Добавить
              </Button>
            </form>

            {/* List 4 States */}
            {tasksQuery.isLoading && (
              <div className="space-y-3 py-2">
                {[1, 2, 3, 4].map((i) => (
                  <div
                    key={i}
                    className="flex items-center gap-3 p-3 rounded-lg border-b border-subtle/40"
                  >
                    <Skeleton variant="circle" className="w-[18px] h-[18px] shrink-0" />
                    <Skeleton variant="text" className="w-3/5 h-4" />
                    <Skeleton variant="rect" className="w-14 h-5 ml-auto" />
                  </div>
                ))}
              </div>
            )}

            {!tasksQuery.isLoading && tasksQuery.isError && (
              <ErrorBanner
                title="Не удалось загрузить задачи проекта"
                message={tasksQuery.error?.message || 'Ошибка сети'}
                onRetry={() => tasksQuery.refetch()}
              />
            )}

            {!tasksQuery.isLoading && !tasksQuery.isError && tasks.length === 0 && (
              <EmptyState
                icon={<FolderClosed className="w-7 h-7 stroke-[1.5]" />}
                title="В проекте нет задач"
                description="Добавьте задачу через поле выше или переключитесь на Канбан-доску."
              />
            )}

            {!tasksQuery.isLoading && !tasksQuery.isError && tasks.length > 0 && visibleTasks.length === 0 && (
              <EmptyState
                icon={<Filter className="w-7 h-7 stroke-[1.5]" />}
                title="Задачи не найдены"
                description="Нет задач, соответствующих выбранным фильтрам. Попробуйте сбросить фильтры."
              />
            )}

            {!tasksQuery.isLoading && !tasksQuery.isError && visibleTasks.length > 0 && (
              <div className="space-y-0.5">
                {visibleTasks.map((task) => (
                  <TaskRow
                    key={task.id}
                    task={task}
                    isSelected={selectedTaskId === task.id}
                    onSelect={() => onSelectTask?.(task.id)}
                    onToggle={(id) => toggleTaskMutation.mutate(id)}
                    onDelete={(id) => deleteTaskMutation.mutate(id)}
                  />
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
