import React, { useState, useMemo } from 'react';
import { Search, FolderKanban, Check } from 'lucide-react';
import { Modal, Input, Badge, EmptyState, Skeleton } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';

export interface BoardColumnOption {
  id: string;
  name: string;
}

export interface BoardGroupOption {
  boardId: string;
  boardName: string;
  columns: BoardColumnOption[];
}

export interface MoveTaskModalProps {
  isOpen: boolean;
  onClose: () => void;
  taskTitle: string;
  currentBoardId?: string | null;
  currentColumnId?: string | null;
  boardGroups: BoardGroupOption[];
  onMove: (targetBoardId: string, targetColumnId: string) => void;
  isLoading?: boolean;
}

export const MoveTaskModal: React.FC<MoveTaskModalProps> = ({
  isOpen,
  onClose,
  taskTitle,
  currentColumnId,
  boardGroups,
  onMove,
  isLoading = false,
}) => {
  const [searchQuery, setSearchQuery] = useState('');

  const cleanQuery = searchQuery.trim().toLowerCase();

  const filteredGroups = useMemo(() => {
    if (!cleanQuery) return boardGroups;

    return boardGroups
      .map((bg) => {
        const boardMatches = bg.boardName.toLowerCase().includes(cleanQuery);
        const matchingColumns = bg.columns.filter((c) =>
          c.name.toLowerCase().includes(cleanQuery)
        );

        if (boardMatches) {
          return bg;
        }

        if (matchingColumns.length > 0) {
          return {
            ...bg,
            columns: matchingColumns,
          };
        }

        return null;
      })
      .filter((bg): bg is BoardGroupOption => bg !== null);
  }, [boardGroups, cleanQuery]);

  const handleSelectColumn = (boardId: string, columnId: string) => {
    if (columnId === currentColumnId) return;
    onMove(boardId, columnId);
    onClose();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Переместить задачу"
      description={`Выберите целевую доску и колонку для «${taskTitle}»`}
      className="max-w-md"
    >
      <div className="space-y-3">
        {/* Search Input with AutoFocus */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted pointer-events-none" />
          <Input
            autoFocus
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Поиск доски или колонки..."
            className="pl-9 h-9 text-xs bg-canvas/60 border-subtle focus-within:bg-surface"
          />
        </div>

        {/* Boards & Columns List */}
        <div className="max-h-72 overflow-y-auto pr-1 space-y-3">
          {isLoading ? (
            <div className="space-y-2 py-2">
              <Skeleton className="h-6 w-1/3 rounded" />
              <Skeleton className="h-8 w-full rounded-md" />
              <Skeleton className="h-8 w-full rounded-md" />
            </div>
          ) : filteredGroups.length === 0 ? (
            <EmptyState
              title={cleanQuery ? 'Ничего не найдено' : 'Нет доступных колонок'}
              description={
                cleanQuery
                  ? 'Попробуйте изменить поисковый запрос'
                  : 'В этом проекте пока нет других колонок'
              }
              className="py-6"
            />
          ) : (
            filteredGroups.map((bg) => (
              <div key={bg.boardId} className="space-y-1">
                {/* Board Group Header */}
                <div className="flex items-center gap-1.5 px-2 py-1 text-[11px] font-semibold text-muted bg-canvas/50 rounded-md">
                  <FolderKanban className="w-3.5 h-3.5 text-muted shrink-0" />
                  <span className="truncate">{bg.boardName}</span>
                </div>

                {/* Columns Rows */}
                <div className="space-y-0.5 pl-2">
                  {bg.columns.map((c) => {
                    const isCurrent = c.id === currentColumnId;
                    return (
                      <button
                        key={c.id}
                        type="button"
                        disabled={isCurrent}
                        onClick={() => handleSelectColumn(bg.boardId, c.id)}
                        className={cn(
                          'w-full text-left px-2.5 py-1.5 rounded-lg text-xs transition-colors flex items-center justify-between group',
                          isCurrent
                            ? 'text-muted/60 bg-canvas/30 cursor-default'
                            : 'text-primary hover:bg-canvas hover:text-accent-solid active:bg-canvas/80'
                        )}
                      >
                        <span className="font-medium truncate">{c.name}</span>
                        {isCurrent ? (
                          <Badge variant="subtle" size="sm" className="text-[10px] shrink-0">
                            (текущая)
                          </Badge>
                        ) : (
                          <span className="text-[10px] text-muted opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-1 shrink-0">
                            <Check className="w-3 h-3" />
                            Перенести
                          </span>
                        )}
                      </button>
                    );
                  })}
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </Modal>
  );
};
