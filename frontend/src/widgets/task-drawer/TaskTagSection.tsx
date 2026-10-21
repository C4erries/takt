import React, { useState, useRef, useEffect } from 'react';
import { X, Check } from 'lucide-react';
import { Input, Button } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import type { Tag } from '@/shared/api';
import {
  useWorkspaceTagsQuery,
  useCreateTagMutation,
  useAddTagToTaskMutation,
  useRemoveTagFromTaskMutation,
} from '@/entities/task';

export const TAG_PALETTE = [
  { name: 'Деним', hex: '#4A6B82' },
  { name: 'Хурма', hex: '#D97746' },
  { name: 'Матча', hex: '#566B44' },
  { name: 'Солома', hex: '#B8943F' },
  { name: 'Слива', hex: '#8A4F7D' },
  { name: 'Суми', hex: '#5C5C5C' },
];

export interface TaskTagSectionProps {
  taskId: string;
  tags?: Tag[];
  workspaceId?: string;
  isAddingTag?: boolean;
  onToggleAddingTag?: () => void;
  className?: string;
}

export const TaskTagSection: React.FC<TaskTagSectionProps> = ({
  taskId,
  tags = [],
  workspaceId,
  isAddingTag: controlledIsAdding,
  onToggleAddingTag,
  className,
}) => {
  const [internalIsAdding, setInternalIsAdding] = useState(false);
  const isAdding = controlledIsAdding !== undefined ? controlledIsAdding : internalIsAdding;
  const setIsAdding = (val: boolean) => {
    if (onToggleAddingTag) onToggleAddingTag();
    else setInternalIsAdding(val);
  };

  const [query, setQuery] = useState('');
  const [selectedColor, setSelectedColor] = useState(TAG_PALETTE[2].hex); // Default: Matcha
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement | null>(null);

  const { data: allTags = [] } = useWorkspaceTagsQuery(workspaceId);
  const createTagMutation = useCreateTagMutation(workspaceId);
  const addTagMutation = useAddTagToTaskMutation(workspaceId);
  const removeTagMutation = useRemoveTagFromTaskMutation(workspaceId);

  const cleanQuery = query.trim().replace(/^@/, '');
  const assignedTagIds = new Set(tags.map((t) => t.id));

  // Autocomplete suggestions: workspace tags not yet attached to this task
  const matchingTags = allTags.filter(
    (t) =>
      !assignedTagIds.has(t.id) &&
      t.name.toLowerCase().includes(cleanQuery.toLowerCase())
  );

  const exactMatch = allTags.some(
    (t) => t.name.toLowerCase() === cleanQuery.toLowerCase()
  );

  const handleSelectExistingTag = async (tag: Tag) => {
    await addTagMutation.mutateAsync({ taskId, tag });
    setQuery('');
    setIsDropdownOpen(false);
    setIsAdding(false);
  };

  const handleCreateNewTag = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!cleanQuery) return;

    const existing = allTags.find(
      (t) => t.name.toLowerCase() === cleanQuery.toLowerCase()
    );

    if (existing) {
      await addTagMutation.mutateAsync({ taskId, tag: existing });
    } else {
      const created = await createTagMutation.mutateAsync({
        name: cleanQuery,
        colorHex: selectedColor,
      });
      await addTagMutation.mutateAsync({ taskId, tag: created });
    }

    setQuery('');
    setIsDropdownOpen(false);
    setIsAdding(false);
  };

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setIsDropdownOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div className={cn('space-y-2', className)}>
      {/* Existing Tags Capsules (NO # symbol) */}
      <div className="flex items-center gap-1.5 flex-wrap">
        {tags && tags.length > 0 ? (
          tags.map((tag) => (
            <span
              key={tag.id}
              className="inline-flex items-center gap-1.5 bg-accent-surface/80 text-primary text-[11px] font-medium px-2.5 py-0.5 rounded-full border border-subtle transition-colors shadow-washi-sm"
            >
              {tag.colorHex && (
                <span
                  className="w-2 h-2 rounded-full shrink-0 shadow-xs"
                  style={{ backgroundColor: tag.colorHex }}
                />
              )}
              <span>{tag.name}</span>
              <button
                type="button"
                onClick={() => removeTagMutation.mutate({ taskId, tag })}
                className="text-muted hover:text-primary transition-colors ml-0.5 rounded-full p-0.5 hover:bg-canvas/50"
                title="Удалить тег"
              >
                <X className="w-2.5 h-2.5" />
              </button>
            </span>
          ))
        ) : (
          !isAdding && <span className="text-[11px] text-muted italic">Нет тегов</span>
        )}
      </div>

      {/* Add Tag Section with Live Autocomplete & Palette */}
      {isAdding && (
        <div ref={dropdownRef} className="relative pt-1 space-y-2">
          <form onSubmit={handleCreateNewTag} className="flex items-center gap-1.5">
            <div className="relative flex-1">
              <Input
                autoFocus
                placeholder="@контекст или название..."
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value);
                  setIsDropdownOpen(true);
                }}
                onFocus={() => setIsDropdownOpen(true)}
                className="text-xs bg-surface border-subtle h-8 w-full"
              />
            </div>
            <Button type="submit" size="sm" variant="solid" className="h-8 text-xs px-3 shrink-0">
              {exactMatch ? 'Добавить' : 'Создать'}
            </Button>
            <Button
              type="button"
              size="sm"
              variant="ghost"
              onClick={() => setIsAdding(false)}
              className="h-8 text-xs px-2 shrink-0"
              title="Отмена"
            >
              ✕
            </Button>
          </form>

          {/* Autocomplete Dropdown & Palette */}
          {isDropdownOpen && (
            <div className="absolute left-0 right-0 top-full mt-1 bg-surface border border-subtle rounded-xl shadow-washi-md z-50 p-2 text-xs space-y-2">
              {/* Existing Matching Tags Suggestions */}
              {matchingTags.length > 0 && (
                <div className="space-y-1">
                  <span className="text-[10px] font-mono uppercase tracking-wider text-muted px-1">
                    Существующие теги
                  </span>
                  <div className="max-h-32 overflow-y-auto space-y-0.5">
                    {matchingTags.map((tag) => (
                      <button
                        key={tag.id}
                        type="button"
                        onClick={() => handleSelectExistingTag(tag)}
                        className="w-full text-left px-2.5 py-1.5 rounded-lg hover:bg-canvas/70 text-primary transition-colors flex items-center justify-between group"
                      >
                        <div className="flex items-center gap-2">
                          {tag.colorHex && (
                            <span
                              className="w-2.5 h-2.5 rounded-full shrink-0"
                              style={{ backgroundColor: tag.colorHex }}
                            />
                          )}
                          <span className="font-medium">{tag.name}</span>
                        </div>
                        <span className="text-[10px] text-muted opacity-0 group-hover:opacity-100 transition-opacity">
                          Выбрать ↵
                        </span>
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Create New Tag with Pastel Color Palette */}
              {cleanQuery && !exactMatch && (
                <div className="pt-1 border-t border-subtle/60 space-y-2">
                  <div className="flex items-center justify-between px-1">
                    <span className="text-[11px] text-primary font-medium">
                      Новый тег: <span className="text-accent-solid">«{cleanQuery}»</span>
                    </span>
                  </div>

                  {/* 6 Pastel Colors Palette (D-071-2) */}
                  <div className="space-y-1 px-1">
                    <span className="text-[10px] font-mono uppercase tracking-wider text-muted">
                      Цвет метки
                    </span>
                    <div className="flex items-center gap-2 pt-0.5">
                      {TAG_PALETTE.map((c) => {
                        const isSelected = selectedColor === c.hex;
                        return (
                          <button
                            key={c.hex}
                            type="button"
                            title={c.name}
                            onClick={() => setSelectedColor(c.hex)}
                            style={{ backgroundColor: c.hex }}
                            className={cn(
                              'w-5 h-5 rounded-full flex items-center justify-center transition-transform hover:scale-110',
                              isSelected
                                ? 'ring-2 ring-primary ring-offset-2 ring-offset-surface scale-105'
                                : 'opacity-80 hover:opacity-100'
                            )}
                          >
                            {isSelected && <Check className="w-3 h-3 text-white drop-shadow-xs" />}
                          </button>
                        );
                      })}
                    </div>
                  </div>

                  <button
                    type="button"
                    onClick={() => handleCreateNewTag()}
                    className="w-full mt-1.5 py-1.5 px-3 rounded-lg bg-accent-solid text-white text-xs font-medium hover:bg-accent-solid-hover transition-colors text-center"
                  >
                    ✨ Создать тег «{cleanQuery}»
                  </button>
                </div>
              )}

              {matchingTags.length === 0 && !cleanQuery && (
                <p className="px-2 py-1 text-[11px] text-muted italic">
                  Начните вводить название тега...
                </p>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
};
