import React, { useState, useMemo } from 'react';
import { FolderClosed, Plus, Filter } from 'lucide-react';
import { ViewToggle, KanbanBoard } from '@/features/kanban-board';
import { useTaskStore } from '@/features/task-navigation';
import {
  useProjectsQuery,
  useProjectTasksQuery,
} from '@/entities/project';
import { useToggleTaskMutation, useDeleteTaskMutation } from '@/features/task-navigation';
import { BoardFilterBar, filterAndSortTasks, initialFilterState } from '@/features/board-filters';
import type { BoardFilterState } from '@/features/board-filters';
import { TaskRow } from './TaskRow';
import { Skeleton, EmptyState, ErrorBanner, Input, Button } from '@/shared/components/ui';

export interface ProjectViewProps {
  projectId: string;
  workspaceId?: string;
  onSelectTask?: (taskId: string) => void;
  selectedTaskId?: string | null;
}

export const ProjectView: React.FC<ProjectViewProps> = ({
  projectId,
  workspaceId,
  onSelectTask,
  selectedTaskId,
}) => {
  const { viewMode, setViewMode } = useTaskStore();
  const { data: projects = [] } = useProjectsQuery(workspaceId);
  const currentProject = projects.find((p) => p.id === projectId) ?? projects[0];

  const tasksQuery = useProjectTasksQuery(workspaceId, projectId);
  const tasks = tasksQuery.data ?? [];

  const [filters, setFilters] = useState<BoardFilterState>(initialFilterState);
  const visibleTasks = useMemo(() => filterAndSortTasks(tasks, filters), [tasks, filters]);

  const toggleTaskMutation = useToggleTaskMutation();
  const deleteTaskMutation = useDeleteTaskMutation();

  const [inlineTitle, setInlineTitle] = useState('');

  const handleInlineSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inlineTitle.trim()) return;
    setInlineTitle('');
  };


  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Project Header */}
      <div className="p-8 pb-4 border-b border-subtle/50 space-y-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 text-xs font-mono uppercase tracking-wider text-muted">
            <span className="w-2 h-2 rounded-full bg-accent-solid inline-block" />
            <span>Проект · {currentProject?.name || 'Takt'}</span>
          </div>

          {/* View Toggle */}
          <ViewToggle view={viewMode} onViewChange={setViewMode} />
        </div>

        <div className="flex items-center justify-between">
          <div>
            <h1 className="font-serif text-3xl font-normal tracking-tight text-primary">
              {currentProject?.name || 'Проект'}
            </h1>
            <p className="text-xs text-muted mt-0.5">
              {currentProject?.description || 'Управление структурой задач и ходом выполнения'}
            </p>
          </div>
        </div>
      </div>

      {/* Project Content: Kanban vs List */}
      <div className="flex-1 overflow-y-auto p-6 space-y-4">
        {/* Board Multi-Filters & View Sorting */}
        <BoardFilterBar
          filters={filters}
          onFilterChange={setFilters}
          workspaceId={workspaceId}
        />

        {viewMode === 'kanban' ? (
          <KanbanBoard
            projectId={projectId}
            workspaceId={workspaceId}
            selectedTaskId={selectedTaskId}
            onSelectTask={onSelectTask}
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
