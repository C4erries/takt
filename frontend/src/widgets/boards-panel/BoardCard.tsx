import React, { useState, useRef, useEffect } from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { MoreHorizontal, Pencil, Trash2, Check, X, GripVertical } from 'lucide-react';
import { Button } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import type { Board } from '@/entities/project';

export interface BoardCardProps {
  board: Board;
  taskCount: number;
  columnCount?: number;
  isActive: boolean;
  canDelete: boolean;
  onSelect: () => void;
  onRename?: (newName: string) => void;
  onDelete?: () => void;
}

function formatColumns(n: number): string {
  const mod10 = n % 10;
  const mod100 = n % 100;
  if (mod100 >= 11 && mod100 <= 14) return `${n} колонок`;
  if (mod10 === 1) return `${n} колонка`;
  if (mod10 >= 2 && mod10 <= 4) return `${n} колонки`;
  return `${n} колонок`;
}

function formatTasks(n: number): string {
  const mod10 = n % 10;
  const mod100 = n % 100;
  if (mod100 >= 11 && mod100 <= 14) return `${n} задач`;
  if (mod10 === 1) return `${n} задача`;
  if (mod10 >= 2 && mod10 <= 4) return `${n} задачи`;
  return `${n} задач`;
}

export const BoardCard: React.FC<BoardCardProps> = ({
  board,
  taskCount,
  columnCount = 0,
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

  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: board.id,
    disabled: isEditing,
  });

  const style: React.CSSProperties = {
    transform: CSS.Translate.toString(transform),
    transition,
  };

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
        onPointerDown={(e) => e.stopPropagation()}
        onClick={(e) => e.stopPropagation()}
        className="p-2.5 rounded-xl bg-surface border border-accent-solid/60 shadow-washi flex items-center gap-1.5"
      >
        <input
          type="text"
          autoFocus
          value={editedName}
          onChange={(e) => setEditedName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Escape') handleCancelRename();
          }}
          className="flex-1 min-w-0 px-2 py-1 text-xs bg-surface border border-subtle rounded-md text-primary focus:outline-none focus:border-accent-solid"
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
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      onClick={onSelect}
      className={cn(
        'group relative flex flex-col justify-between p-2.5 min-h-[58px] rounded-xl border transition-all duration-150 cursor-grab active:cursor-grabbing select-none text-xs',
        isDragging && 'opacity-30 border-dashed border-subtle/80 shadow-none',
        isActive
          ? 'bg-surface border-accent-solid/70 shadow-washi-sm text-primary'
          : 'border-transparent bg-surface/30 text-muted hover:text-primary hover:bg-surface/70 hover:border-subtle/50'
      )}
    >
      <div className="flex items-center justify-between gap-1 w-full">
        <div className="flex items-center gap-1.5 min-w-0 pr-1">
          <div className="text-muted/30 group-hover:text-muted/70 transition-colors shrink-0 -ml-0.5">
            <GripVertical className="w-3.5 h-3.5" />
          </div>
          <span
            className={cn(
              'w-2 h-2 rounded-full shrink-0 transition-colors',
              isActive ? 'bg-accent-solid' : 'bg-muted/40 group-hover:bg-muted'
            )}
          />
          <span className="font-medium text-sm text-primary truncate">{board.name}</span>
        </div>

        <div className="relative shrink-0" ref={menuRef} onPointerDown={(e) => e.stopPropagation()}>
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

      {/* Second row: board metrics */}
      <div className="flex items-center gap-1.5 text-[11px] text-muted font-mono mt-1 pl-4">
        <span>⊞ {formatColumns(columnCount)}</span>
        <span className="text-muted/50">·</span>
        <span>✓ {formatTasks(taskCount)}</span>
      </div>
    </div>
  );
};
