import React, { useState, useRef, useEffect } from 'react';
import { MoreHorizontal, Pencil, Trash2, Check, X } from 'lucide-react';
import { Badge, Button } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import type { Board } from '@/entities/project';

export interface BoardCardProps {
  board: Board;
  taskCount: number;
  isActive: boolean;
  canDelete: boolean;
  onSelect: () => void;
  onRename?: (newName: string) => void;
  onDelete?: () => void;
}

export const BoardCard: React.FC<BoardCardProps> = ({
  board,
  taskCount,
  isActive,
  canDelete,
  onSelect,
  onRename,
  onDelete,
}) => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editedName, setEditedName] = useState(board.name);
  const menuRef = useRef<HTMLDivElement>(null);

  // Close menu on click outside
  useEffect(() => {
    if (!isMenuOpen) return;
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setIsMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isMenuOpen]);

  const handleSaveRename = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const trimmed = editedName.trim();
    if (trimmed && trimmed !== board.name && onRename) {
      onRename(trimmed);
    }
    setIsEditing(false);
    setIsMenuOpen(false);
  };

  const handleCancelRename = () => {
    setEditedName(board.name);
    setIsEditing(false);
  };

  const handleDelete = () => {
    setIsMenuOpen(false);
    if (!canDelete) return;
    if (window.confirm(`Удалить доску «${board.name}»?`)) {
      onDelete?.();
    }
  };

  if (isEditing) {
    return (
      <form
        onSubmit={handleSaveRename}
        className="p-2 rounded-xl bg-surface border border-accent-solid/60 shadow-washi flex items-center gap-1.5"
      >
        <input
          type="text"
          autoFocus
          value={editedName}
          onChange={(e) => setEditedName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Escape') handleCancelRename();
          }}
          className="flex-1 min-w-0 px-2 py-1 text-xs bg-canvas/40 border border-subtle rounded-md text-primary focus:outline-none focus:border-accent-solid"
        />
        <Button
          type="submit"
          variant="ghost"
          size="sm"
          className="h-6 w-6 p-0 text-accent-solid hover:bg-accent-surface/30"
          title="Сохранить"
        >
          <Check className="w-3.5 h-3.5" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={handleCancelRename}
          className="h-6 w-6 p-0 text-muted hover:text-primary"
          title="Отмена"
        >
          <X className="w-3.5 h-3.5" />
        </Button>
      </form>
    );
  }

  return (
    <div
      onClick={onSelect}
      className={cn(
        'group relative flex items-center justify-between px-3 py-2 rounded-xl border transition-all duration-150 cursor-pointer select-none text-xs',
        isActive
          ? 'bg-surface border-subtle shadow-washi text-primary font-medium'
          : 'border-transparent text-muted hover:text-primary hover:bg-surface/50 hover:border-subtle/40'
      )}
    >
      <div className="flex items-center gap-2 min-w-0 pr-1">
        <span
          className={cn(
            'w-1.5 h-1.5 rounded-full shrink-0 transition-colors',
            isActive ? 'bg-accent-solid' : 'bg-muted/40 group-hover:bg-muted'
          )}
        />
        <span className="truncate">{board.name}</span>
      </div>

      <div className="flex items-center gap-1 shrink-0">
        <Badge
          variant={isActive ? 'subtle' : 'subtle'}
          size="sm"
          className={cn(
            'font-mono text-[10px] px-1.5 py-0 h-4.5',
            isActive
              ? 'bg-canvas text-primary font-semibold'
              : 'text-muted bg-canvas/60'
          )}
        >
          {taskCount}
        </Badge>

        <div className="relative" ref={menuRef}>
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              setIsMenuOpen((prev) => !prev);
            }}
            className={cn(
              'p-1 rounded-md text-muted hover:text-primary transition-opacity',
              isMenuOpen ? 'opacity-100 bg-canvas/60' : 'opacity-0 group-hover:opacity-100'
            )}
            title="Действия с доской"
          >
            <MoreHorizontal className="w-3.5 h-3.5" />
          </button>

          {isMenuOpen && (
            <div
              onClick={(e) => e.stopPropagation()}
              className="absolute right-0 top-full mt-1 w-40 z-30 bg-surface border border-subtle rounded-xl shadow-washi p-1 space-y-0.5 text-xs animate-in fade-in zoom-in-95 duration-100"
            >
              <button
                type="button"
                onClick={() => {
                  setIsMenuOpen(false);
                  setIsEditing(true);
                }}
                className="w-full flex items-center gap-2 px-2.5 py-1.5 rounded-lg text-left text-primary hover:bg-canvas/60 transition-colors"
              >
                <Pencil className="w-3 h-3 text-muted" />
                <span>Переименовать</span>
              </button>

              <button
                type="button"
                disabled={!canDelete}
                onClick={handleDelete}
                className={cn(
                  'w-full flex items-center gap-2 px-2.5 py-1.5 rounded-lg text-left transition-colors',
                  canDelete
                    ? 'text-primary hover:bg-canvas/60'
                    : 'text-muted/50 cursor-not-allowed'
                )}
                title={!canDelete ? 'Нельзя удалить единственную доску' : undefined}
              >
                <Trash2 className="w-3 h-3 text-muted" />
                <span>Удалить</span>
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
