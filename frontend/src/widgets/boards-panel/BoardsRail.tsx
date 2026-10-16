import React, { useState } from 'react';
import { ChevronLeft, ChevronRight, Plus, X, LayoutGrid, Check } from 'lucide-react';
import { useQueries } from '@tanstack/react-query';
import { getAuthToken } from '@/shared/api';
import {
  useProjectBoardsQuery,
  useProjectTasksQuery,
  useCreateBoardMutation,
  DEMO_COLUMNS,
  projectApi,
  type Column,
} from '@/entities/project';
import { Skeleton, EmptyState, ErrorBanner, Button } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import { BoardCard } from './BoardCard';

export interface BoardsRailProps {
  workspaceId?: string;
  projectId: string;
  activeBoardId?: string | null;
  onSelectBoard: (boardId: string) => void;
  isCollapsed: boolean;
  onToggleCollapse: () => void;
  className?: string;
}

export const BoardsRail: React.FC<BoardsRailProps> = ({
  workspaceId,
  projectId,
  activeBoardId,
  onSelectBoard,
  isCollapsed,
  onToggleCollapse,
  className,
}) => {
  const [isCreating, setIsCreating] = useState(false);
  const [newBoardName, setNewBoardName] = useState('');

  const boardsQuery = useProjectBoardsQuery(workspaceId, projectId);
  const boards = boardsQuery.data ?? [];

  const tasksQuery = useProjectTasksQuery(workspaceId, projectId);
  const tasks = tasksQuery.data ?? [];

  const createBoardMutation = useCreateBoardMutation(workspaceId, projectId);

  // Pre-fetch columns for each board to count tasks per board
  const columnQueries = useQueries({
    queries: boards.map((b) => ({
      queryKey: ['columns', workspaceId, b.id],
      queryFn: async (): Promise<Column[]> => {
        if (!workspaceId || !getAuthToken()) {
          const matching = DEMO_COLUMNS.filter((c) => c.boardId === b.id);
          return matching.length > 0 ? matching : DEMO_COLUMNS;
        }
        return projectApi.listBoardColumns(workspaceId, b.id);
      },
      enabled: !!workspaceId && !!b.id,
    })),
  });

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = newBoardName.trim();
    if (!trimmed || createBoardMutation.isPending) return;

    try {
      const created = await createBoardMutation.mutateAsync(trimmed);
      if (created?.id) {
        onSelectBoard(created.id);
      }
      setNewBoardName('');
      setIsCreating(false);
    } catch {
      // Error handled by react-query
    }
  };

  // If collapsed, render 0-width rail with floating expand button to grant 100% canvas to kanban
  if (isCollapsed) {
    return (
      <div className="relative z-20 shrink-0">
        <button
          type="button"
          onClick={onToggleCollapse}
          className="absolute top-4 left-1 p-1.5 rounded-lg bg-surface/90 hover:bg-surface border border-subtle shadow-washi text-muted hover:text-primary transition-all duration-150 flex items-center justify-center group"
          title="Развернуть панели досок (>)"
          aria-label="Развернуть панели досок"
        >
          <ChevronRight className="w-4 h-4 text-muted group-hover:text-primary transition-colors" />
        </button>
      </div>
    );
  }

  const isLoading = boardsQuery.isLoading;
  const isError = boardsQuery.isError;

  return (
    <aside
      className={cn(
        'w-[220px] shrink-0 h-full flex flex-col py-3 pr-2 pl-3 transition-all duration-200 select-none border-r border-subtle/50',
        className
      )}
    >
      {/* Rail Header */}
      <div className="flex items-center justify-between pb-3 px-1 border-b border-subtle/40">
        <div className="flex items-center gap-1.5 text-xs font-semibold tracking-wider uppercase text-muted">
          <LayoutGrid className="w-3.5 h-3.5 text-accent-solid" />
          <span>Доски</span>
        </div>

        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsCreating((prev) => !prev)}
            className="h-6 w-6 p-0 text-muted hover:text-primary"
            title="Добавить доску"
            aria-label="Добавить доску"
          >
            {isCreating ? <X className="w-3.5 h-3.5" /> : <Plus className="w-3.5 h-3.5" />}
          </Button>

          <Button
            variant="ghost"
            size="sm"
            onClick={onToggleCollapse}
            className="h-6 w-6 p-0 text-muted hover:text-primary"
            title="Свернуть панели досок (<)"
            aria-label="Свернуть панели досок"
          >
            <ChevronLeft className="w-3.5 h-3.5" />
          </Button>
        </div>
      </div>

      {/* Inline Create Board Form */}
      {isCreating && (
        <form onSubmit={handleCreateSubmit} className="mt-2.5 p-2 rounded-xl bg-surface border border-subtle shadow-washi space-y-2">
          <input
            type="text"
            autoFocus
            value={newBoardName}
            onChange={(e) => setNewBoardName(e.target.value)}
            placeholder="Название доски..."
            disabled={createBoardMutation.isPending}
            className="w-full px-2 py-1 text-xs bg-canvas/40 border border-subtle rounded-md text-primary placeholder:text-muted/60 focus:outline-none focus:border-accent-solid"
          />
          <div className="flex items-center justify-end gap-1.5">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => {
                setIsCreating(false);
                setNewBoardName('');
              }}
              className="text-[11px] h-6 px-2 text-muted"
            >
              Отмена
            </Button>
            <Button
              type="submit"
              variant="solid"
              size="sm"
              disabled={!newBoardName.trim()}
              isLoading={createBoardMutation.isPending}
              className="text-[11px] h-6 px-2.5"
            >
              <Check className="w-3 h-3 mr-1" />
              Создать
            </Button>
          </div>
        </form>
      )}

      {/* Content: 4 States */}
      <div className="flex-1 overflow-y-auto pt-2 space-y-1">
        {/* State 1: Loading Skeleton */}
        {isLoading && (
          <div className="space-y-2 pt-1">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex items-center justify-between p-2.5 rounded-xl border border-subtle/30 bg-surface/30">
                <Skeleton variant="text" className="w-24 h-3.5" />
                <Skeleton variant="circle" className="w-4 h-4" />
              </div>
            ))}
          </div>
        )}

        {/* State 2: Error */}
        {!isLoading && isError && (
          <div className="py-4">
            <ErrorBanner
              title="Ошибка досок"
              message={boardsQuery.error?.message || 'Не удалось загрузить доски'}
              onRetry={() => boardsQuery.refetch()}
            />
          </div>
        )}

        {/* State 3: Empty */}
        {!isLoading && !isError && boards.length === 0 && (
          <div className="py-6">
            <EmptyState
              icon={<LayoutGrid className="w-6 h-6 stroke-[1.5]" />}
              title="Нет досок"
              description="Создайте первую доску проекта."
              action={
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => setIsCreating(true)}
                  className="text-xs"
                >
                  <Plus className="w-3 h-3 mr-1" /> Добавить
                </Button>
              }
            />
          </div>
        )}

        {/* State 4: Success - Boards List */}
        {!isLoading &&
          !isError &&
          boards.map((board, idx) => {
            const boardCols = columnQueries[idx]?.data ?? [];
            const colIds = new Set(boardCols.map((c) => c.id));
            const count = tasks.filter((t) => t.columnId && colIds.has(t.columnId)).length;
            const isActive = activeBoardId ? board.id === activeBoardId : idx === 0;

            return (
              <BoardCard
                key={board.id}
                board={board}
                taskCount={count}
                isActive={isActive}
                canDelete={boards.length > 1}
                onSelect={() => onSelectBoard(board.id)}
              />
            );
          })}
      </div>
    </aside>
  );
};
