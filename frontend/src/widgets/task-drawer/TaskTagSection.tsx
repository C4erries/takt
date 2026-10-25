import React, { useState, useRef, useEffect, useMemo } from 'react';
import { Check, Folder, Globe } from 'lucide-react';
import { Input, Button } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import type { Tag } from '@/shared/api';
import {
  TAG_PALETTE,
  DEFAULT_TAG_COLOR,
  useWorkspaceTagsQuery,
  useCreateTagMutation,
  useUpdateTagMutation,
  useAddTagToTaskMutation,
  useRemoveTagFromTaskMutation,
} from '@/entities/task';
import { TagCapsule } from './TagCapsule';

export interface TaskTagSectionProps {
  taskId: string;
  tags?: Tag[];
  workspaceId?: string;
  projectId?: string | null;
  isAddingTag?: boolean;
  onToggleAddingTag?: () => void;
  className?: string;
}

export const TaskTagSection: React.FC<TaskTagSectionProps> = ({
  taskId,
  tags = [],
  workspaceId,
  projectId,
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
  const [selectedColor, setSelectedColor] = useState(DEFAULT_TAG_COLOR);
  const [createScope, setCreateScope] = useState<'project' | 'global'>(projectId ? 'project' : 'global');
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const [activeColorPopoverTagId, setActiveColorPopoverTagId] = useState<string | null>(null);
  const dropdownRef = useRef<HTMLDivElement | null>(null);
  const itemRefs = useRef<(HTMLButtonElement | null)[]>([]);

  // Fetch tags: will return global tags and tags of this project (if projectId is given)
  const { data: allTags = [] } = useWorkspaceTagsQuery(workspaceId, projectId);
  const createTagMutation = useCreateTagMutation(workspaceId);
  const updateTagMutation = useUpdateTagMutation(workspaceId);
  const addTagMutation = useAddTagToTaskMutation(workspaceId);
  const removeTagMutation = useRemoveTagFromTaskMutation(workspaceId);

  const cleanQuery = query.trim().replace(/^@/, '');
  const assignedTagIds = new Set(tags.map((t) => t.id));

  // Filter out tags not relevant to this task:
  // Show global tags + tags belonging to this task's project
  const relevantTags = useMemo(() => {
    return allTags.filter((t) => {
      if (!t.projectId) return true; // Global
      if (projectId && t.projectId === projectId) return true; // Matches task's project
      return false; // Other project's tag
    });
  }, [allTags, projectId]);

  // Autocomplete suggestions: tags not yet attached to this task
  const matchingTags = useMemo(() => {
    return relevantTags.filter(
      (t) =>
        !assignedTagIds.has(t.id) &&
        t.name.toLowerCase().includes(cleanQuery.toLowerCase())
    );
  }, [relevantTags, assignedTagIds, cleanQuery]);

  const exactMatch = useMemo(() => {
    return relevantTags.some(
      (t) => t.name.toLowerCase() === cleanQuery.toLowerCase()
    );
  }, [relevantTags, cleanQuery]);

  const resetInput = () => {
    setQuery('');
    setFocusedIndex(-1);
    setIsDropdownOpen(false);
    setIsAdding(false);
  };

  const handleSelectExistingTag = async (tag: Tag) => {
    await addTagMutation.mutateAsync({ taskId, tag });
    resetInput();
  };

  const handleCreateNewTag = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!cleanQuery) return;

    const existing = relevantTags.find(
      (t) => t.name.toLowerCase() === cleanQuery.toLowerCase()
    );

    if (existing) {
      await addTagMutation.mutateAsync({ taskId, tag: existing });
    } else {
      const targetProjectId = createScope === 'project' && projectId ? projectId : null;
      const created = await createTagMutation.mutateAsync({
        name: cleanQuery,
        colorHex: selectedColor,
        projectId: targetProjectId,
      });
      await addTagMutation.mutateAsync({ taskId, tag: created });
    }
    resetInput();
  };

  useEffect(() => {
    if (focusedIndex >= 0 && itemRefs.current[focusedIndex]) {
      itemRefs.current[focusedIndex]?.scrollIntoView({ block: 'nearest' });
    }
  }, [focusedIndex]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Escape') {
      e.preventDefault();
      setIsDropdownOpen(false);
      setFocusedIndex(-1);
      return;
    }

    if (!isDropdownOpen || matchingTags.length === 0) return;

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setFocusedIndex((prev) => (prev + 1) % matchingTags.length);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setFocusedIndex((prev) => (prev - 1 + matchingTags.length) % matchingTags.length);
    } else if (e.key === 'Enter') {
      if (focusedIndex >= 0 && matchingTags[focusedIndex]) {
        e.preventDefault();
        handleSelectExistingTag(matchingTags[focusedIndex]);
      }
    }
  };

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as Node;
      if (dropdownRef.current && !dropdownRef.current.contains(target)) {
        setIsDropdownOpen(false);
        setFocusedIndex(-1);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div className={cn('space-y-2', className)}>
      {/* Existing Tags Capsules */}
      <div className="flex items-center gap-1.5 flex-wrap">
        {tags && tags.length > 0 ? (
          tags.map((tag) => (
            <TagCapsule
              key={tag.id}
              tag={tag}
              isOpen={activeColorPopoverTagId === tag.id}
              onToggleOpen={() =>
                setActiveColorPopoverTagId(
                  activeColorPopoverTagId === tag.id ? null : tag.id
                )
              }
              onClose={() => setActiveColorPopoverTagId(null)}
              onUpdateColor={(hex) =>
                updateTagMutation.mutate({ tagId: tag.id, colorHex: hex })
              }
              onRemove={() => removeTagMutation.mutate({ taskId, tag })}
            />
          ))
        ) : (
          !isAdding && <span className="text-[11px] text-muted italic">Нет тегов</span>
        )}
      </div>

      {/* Add Tag Section with Live Autocomplete & Scope Choice */}
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
                  setFocusedIndex(-1);
                  setIsDropdownOpen(true);
                }}
                onKeyDown={handleKeyDown}
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
              onClick={() => {
                setIsAdding(false);
                setFocusedIndex(-1);
              }}
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
                    {matchingTags.map((tag, idx) => {
                      const isFocused = focusedIndex === idx;
                      return (
                        <button
                          key={tag.id}
                          ref={(el) => {
                            itemRefs.current[idx] = el;
                          }}
                          type="button"
                          onClick={() => handleSelectExistingTag(tag)}
                          className={cn(
                            'w-full text-left px-2.5 py-1.5 rounded-lg text-primary transition-colors flex items-center justify-between group',
                            isFocused ? 'bg-canvas/90 ring-1 ring-subtle' : 'hover:bg-canvas/70'
                          )}
                        >
                          <div className="flex items-center gap-2">
                            {tag.colorHex && (
                              <span
                                className="w-2.5 h-2.5 rounded-full shrink-0"
                                style={{ backgroundColor: tag.colorHex }}
                              />
                            )}
                            <span className="font-medium">{tag.name}</span>
                            {tag.projectId ? (
                              <span className="text-[10px] font-mono text-muted/70 flex items-center gap-0.5">
                                <Folder className="w-2.5 h-2.5 text-accent-solid" />
                                проект
                              </span>
                            ) : (
                              <span className="text-[10px] font-mono text-muted/70 flex items-center gap-0.5">
                                <Globe className="w-2.5 h-2.5" />
                                глобальный
                              </span>
                            )}
                          </div>
                          <span
                            className={cn(
                              'text-[10px] text-muted transition-opacity',
                              isFocused ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'
                            )}
                          >
                            Выбрать ↵
                          </span>
                        </button>
                      );
                    })}
                  </div>
                </div>
              )}

              {/* Create New Tag with Scope & 12 Shades Palette */}
              {cleanQuery && !exactMatch && (
                <div className="pt-1 border-t border-subtle/60 space-y-2">
                  <div className="flex items-center justify-between px-1">
                    <span className="text-[11px] text-primary font-medium">
                      Новый тег: <span className="text-accent-solid">«{cleanQuery}»</span>
                    </span>
                    {projectId && (
                      <div className="flex items-center gap-1 text-[10px] font-mono">
                        <button
                          type="button"
                          onClick={() => setCreateScope('project')}
                          className={cn(
                            'px-1.5 py-0.5 rounded transition-colors',
                            createScope === 'project'
                              ? 'bg-accent-surface text-primary font-semibold'
                              : 'text-muted hover:text-primary'
                          )}
                        >
                          Проект
                        </button>
                        <span className="text-muted/40">/</span>
                        <button
                          type="button"
                          onClick={() => setCreateScope('global')}
                          className={cn(
                            'px-1.5 py-0.5 rounded transition-colors',
                            createScope === 'global'
                              ? 'bg-accent-surface text-primary font-semibold'
                              : 'text-muted hover:text-primary'
                          )}
                        >
                          Глобальный
                        </button>
                      </div>
                    )}
                  </div>

                  {/* 12 Color Palette */}
                  <div className="space-y-1 px-1">
                    <span className="text-[10px] font-mono uppercase tracking-wider text-muted">
                      Цвет метки
                    </span>
                    <div className="grid grid-cols-6 gap-1.5 pt-0.5">
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
                              'w-5 h-5 rounded-full flex items-center justify-center transition-transform hover:scale-110 cursor-pointer',
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
                    ✨ Создать {createScope === 'project' && projectId ? 'проектный' : 'глобальный'} тег «{cleanQuery}»
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
