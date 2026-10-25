import React from 'react';
import { Trash2, Pencil, Check, X, Folder, Globe } from 'lucide-react';
import { Badge } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import { TAG_PALETTE, DEFAULT_TAG_COLOR } from '@/entities/task';
import type { TagWithMetrics } from '@/shared/api';

export interface TagCardProps {
  tag: TagWithMetrics;
  projectName?: string;
  isEditing: boolean;
  editingName: string;
  editInputRef: React.RefObject<HTMLInputElement>;
  onStartRename: (tag: TagWithMetrics) => void;
  onEditingNameChange: (val: string) => void;
  onSaveRename: (tagId: string) => void;
  onCancelRename: () => void;
  onColorSelect: (tagId: string, hex: string) => void;
  onDeleteTag: (tag: TagWithMetrics) => void;
}

function formatTaskCount(count: number): string {
  if (count === 0) return '0 задач';
  const mod10 = count % 10;
  const mod100 = count % 100;
  if (mod100 >= 11 && mod100 <= 14) return `${count} задач`;
  if (mod10 === 1) return `${count} задача`;
  if (mod10 >= 2 && mod10 <= 4) return `${count} задачи`;
  return `${count} задач`;
}

export const TagCard: React.FC<TagCardProps> = ({
  tag,
  projectName,
  isEditing,
  editingName,
  editInputRef,
  onStartRename,
  onEditingNameChange,
  onSaveRename,
  onCancelRename,
  onColorSelect,
  onDeleteTag,
}) => {
  const activeHex = (tag.colorHex || DEFAULT_TAG_COLOR).toLowerCase();

  return (
    <div className="group p-3 rounded-xl border border-subtle/70 bg-surface/60 hover:bg-surface hover:border-subtle transition-all duration-150 flex flex-col justify-between min-h-[130px] max-h-[155px] shadow-washi-sm">
      {/* Top: Tag Title / Rename Form & Actions */}
      <div className="flex items-center justify-between gap-2">
        {isEditing ? (
          <div className="flex items-center gap-1 flex-1">
            <input
              ref={editInputRef}
              type="text"
              value={editingName}
              onChange={(e) => onEditingNameChange(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') onSaveRename(tag.id);
                if (e.key === 'Escape') onCancelRename();
              }}
              className="w-full text-xs bg-canvas border border-accent-solid rounded px-2 py-1 text-primary focus:outline-none"
            />
            <button
              type="button"
              onClick={() => onSaveRename(tag.id)}
              className="p-1 text-accent-solid hover:bg-canvas rounded transition-colors"
              title="Сохранить"
            >
              <Check className="w-3.5 h-3.5" />
            </button>
            <button
              type="button"
              onClick={onCancelRename}
              className="p-1 text-muted hover:text-primary hover:bg-canvas rounded transition-colors"
              title="Отмена"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        ) : (
          <>
            <div className="flex items-center gap-1.5 min-w-0 flex-1">
              <span
                className="w-2.5 h-2.5 rounded-full shrink-0"
                style={{ backgroundColor: tag.colorHex || DEFAULT_TAG_COLOR }}
              />
              <span className="text-xs font-semibold text-primary truncate" title={tag.name}>
                #{tag.name}
              </span>
            </div>
            <div className="flex items-center gap-0.5 opacity-80 group-hover:opacity-100 transition-opacity">
              <button
                type="button"
                onClick={() => onStartRename(tag)}
                className="p-1 rounded text-muted hover:text-primary hover:bg-canvas transition-colors"
                title="Переименовать тег"
              >
                <Pencil className="w-3 h-3" />
              </button>
              <button
                type="button"
                onClick={() => onDeleteTag(tag)}
                className="p-1 rounded text-muted hover:text-destructive hover:bg-canvas transition-colors"
                title="Удалить тег"
              >
                <Trash2 className="w-3 h-3" />
              </button>
            </div>
          </>
        )}
      </div>

      {/* Scope Badge & Task Count */}
      <div className="flex items-center justify-between text-[11px] text-muted font-mono gap-1">
        <span className="inline-flex items-center min-w-0">
          {tag.projectId ? (
            <Badge
              variant="subtle"
              size="sm"
              className="text-[10px] flex items-center gap-1 max-w-[130px] truncate"
              title={`Тег проекта: ${projectName || 'Проект'}`}
            >
              <Folder className="w-2.5 h-2.5 shrink-0 text-accent-solid" />
              <span className="truncate">{projectName || 'Проект'}</span>
            </Badge>
          ) : (
            <Badge
              variant="subtle"
              size="sm"
              className="text-[10px] flex items-center gap-1 text-muted/80"
              title="Глобальный тег воркспейса"
            >
              <Globe className="w-2.5 h-2.5 shrink-0" />
              <span>Глобальный</span>
            </Badge>
          )}
        </span>

        <Badge
          variant="subtle"
          size="sm"
          className={cn('text-[10px] shrink-0', tag.taskCount === 0 && 'opacity-60')}
        >
          {formatTaskCount(tag.taskCount)}
        </Badge>
      </div>

      {/* Bottom: 12-Color 1-Click Palette */}
      <div className="pt-1.5 border-t border-subtle/50 space-y-1">
        <div className="flex items-center justify-between">
          <span className="text-[10px] font-mono text-muted/70 block">Цвет метки:</span>
        </div>
        <div className="flex items-center gap-1.5 flex-wrap">
          {TAG_PALETTE.map((c) => {
            const isSelected = activeHex === c.hex.toLowerCase();
            return (
              <button
                key={c.hex}
                type="button"
                onClick={() => onColorSelect(tag.id, c.hex)}
                title={c.name}
                className={cn(
                  'w-4 h-4 rounded-full transition-transform hover:scale-125 focus:outline-none flex items-center justify-center shrink-0',
                  isSelected
                    ? 'ring-2 ring-primary ring-offset-1 ring-offset-surface scale-110'
                    : 'opacity-80 hover:opacity-100'
                )}
                style={{ backgroundColor: c.hex }}
              >
                {isSelected && <Check className="w-2.5 h-2.5 text-white stroke-[3]" />}
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
};
