import React, { useState } from 'react';
import { ChevronLeft, ChevronRight, Plus, LayoutGrid, Check } from 'lucide-react';
import { useQueries } from '@tanstack/react-query';
import { getAuthToken } from '@/shared/api';
import {
  useProjectBoardsQuery,
  useProjectTasksQuery,
  useCreateBoardMutation,
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

  // Pre-fetch columns for each board to count tasks and columns per board
  const columnQueries = useQueries({
    queries: boards.map((b) => ({
      queryKey: ['columns', workspaceId, b.id],
      queryFn: async (): Promise<Column[]> => {
        if (!workspaceId || !getAuthToken()) {
          return [];
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
        'w-[230px] shrink-0 h-full flex flex-col py-3 pr-2.5 pl-3 transition-all duration-200 select-none bg-canvas/60 border-r border-subtle',
        className
      )}
    >
      {/* Rail Header */}
      <div className="flex items-center justify-between pb-2.5 px-1 border-b border-subtle/50">
        <div className="flex items-center gap-1.5 text-xs font-semibold tracking-wider uppercase text-muted">
          <LayoutGrid className="w-3.5 h-3.5 text-accent-solid" />
          <span>Доски</span>
        </div>

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

      {/* Prominent + Добавить доску button */}
      <div className="pt-2.5 pb-2">
        {isCreating ? (
          <form onSubmit={handleCreateSubmit} className="p-2 rounded-xl bg-surface border border-subtle shadow-washi space-y-2">
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
        ) : (
          <Button
            variant="solid"
            size="sm"
            onClick={() => setIsCreating(true)}
            className="w-full justify-center text-xs h-8 shadow-washi-sm"
            leftIcon={<Plus className="w-3.5 h-3.5" />}
          >
            Добавить доску
          </Button>
        )}
      </div>

      {/* Content: 4 States */}
      <div className="flex-1 overflow-y-auto space-y-1.5 pr-0.5">
        {/* State 1: Loading Skeleton */}
        {isLoading && (
          <div className="space-y-2 pt-1">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex flex-col justify-between p-2.5 min-h-[58px] rounded-xl border border-subtle/40 bg-surface/40 space-y-2">
                <Skeleton variant="text" className="w-28 h-4" />
                <Skeleton variant="text" className="w-20 h-3" />
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
                columnCount={boardCols.length}
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
