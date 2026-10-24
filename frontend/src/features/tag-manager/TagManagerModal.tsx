import React, { useState, useRef, useEffect } from 'react';
import { Tag as TagIcon, Trash2, Pencil, Check, Search } from 'lucide-react';
import {
  Modal,
  Input,
  Button,
  Badge,
  Skeleton,
  EmptyState,
  ErrorBanner,
} from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import { TAG_PALETTE } from '@/entities/task';
import type { TagWithMetrics } from '@/shared/api';
import {
  useWorkspaceTagsMetricsQuery,
  useUpdateWorkspaceTagMutation,
  useDeleteWorkspaceTagMutation,
  useDeleteUnusedTagsMutation,
} from './api';

export interface TagManagerModalProps {
  isOpen: boolean;
  onClose: () => void;
  workspaceId?: string;
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

export const TagManagerModal: React.FC<TagManagerModalProps> = ({
  isOpen,
  onClose,
  workspaceId,
}) => {
  const [search, setSearch] = useState('');
  const [editingTagId, setEditingTagId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');
  const [activeColorPickerTagId, setActiveColorPickerTagId] = useState<string | null>(null);

  const editInputRef = useRef<HTMLInputElement>(null);
  const colorPickerRef = useRef<HTMLDivElement>(null);

  const {
    data: tags = [],
    isLoading,
    isError,
    error,
    refetch,
  } = useWorkspaceTagsMetricsQuery(workspaceId);

  const updateTagMutation = useUpdateWorkspaceTagMutation(workspaceId);
  const deleteTagMutation = useDeleteWorkspaceTagMutation(workspaceId);
  const deleteUnusedMutation = useDeleteUnusedTagsMutation(workspaceId);

  useEffect(() => {
    if (editingTagId) {
      editInputRef.current?.focus();
      editInputRef.current?.select();
    }
  }, [editingTagId]);

  useEffect(() => {
    if (!activeColorPickerTagId) return;
    const onMouseDown = (e: MouseEvent) => {
      if (colorPickerRef.current && !colorPickerRef.current.contains(e.target as Node)) {
        setActiveColorPickerTagId(null);
      }
    };
    document.addEventListener('mousedown', onMouseDown);
    return () => document.removeEventListener('mousedown', onMouseDown);
  }, [activeColorPickerTagId]);

  const unusedTags = tags.filter((t) => t.taskCount === 0);
  const unusedCount = unusedTags.length;

  const filteredTags = tags.filter((t) =>
    t.name.toLowerCase().includes(search.trim().toLowerCase())
  );

  const handleStartRename = (tag: TagWithMetrics) => {
    setEditingTagId(tag.id);
    setEditingName(tag.name);
  };

  const handleSaveRename = (tagId: string) => {
    if (!editingTagId) return;
    const trimmed = editingName.trim();
    setEditingTagId(null);
    if (trimmed && trimmed !== tags.find((t) => t.id === tagId)?.name) {
      updateTagMutation.mutate({ tagId, data: { name: trimmed } });
    }
  };

  const handleColorSelect = (tagId: string, hex: string) => {
    updateTagMutation.mutate({ tagId, data: { colorHex: hex } });
    setActiveColorPickerTagId(null);
  };

  const handleDeleteTag = (tag: TagWithMetrics) => {
    const msg =
      tag.taskCount > 0
        ? `Тег «${tag.name}» будет отвязан от ${formatTaskCount(tag.taskCount)}. Продолжить?`
        : `Удалить неиспользуемый тег «${tag.name}»?`;
    if (window.confirm(msg)) {
      deleteTagMutation.mutate(tag.id);
    }
  };

  const handleDeleteUnused = () => {
    if (unusedCount === 0) return;
    if (window.confirm(`Удалить все неиспользуемые теги (${unusedCount} шт.)?`)) {
      deleteUnusedMutation.mutate();
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Менеджер тегов"
      description="Управление каталогом меток воркспейса, редактирование палитры и очистка неиспользуемых тегов"
      className="max-w-xl"
    >
      <div className="space-y-4">
        {/* Controls: Search & Delete Unused */}
        <div className="flex items-center gap-2 flex-wrap justify-between">
          <div className="relative flex-1 min-w-[180px]">
            <Search className="w-3.5 h-3.5 text-muted absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none" />
            <Input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Фильтр тегов..."
              className="text-xs bg-surface border-subtle h-8 w-full pl-8"
            />
          </div>

          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={unusedCount === 0 || deleteUnusedMutation.isPending}
            onClick={handleDeleteUnused}
            className="text-xs h-8 text-muted hover:text-destructive hover:border-destructive/40 transition-colors shrink-0"
            title={unusedCount === 0 ? 'Нет неиспользуемых тегов' : undefined}
          >
            <Trash2 className="w-3.5 h-3.5 mr-1" />
            Удалить неиспользуемые {unusedCount > 0 && `(${unusedCount})`}
          </Button>
        </div>

        {/* State 1: Loading */}
        {isLoading && (
          <div className="space-y-2 py-2">
            {[1, 2, 3, 4].map((i) => (
              <div
                key={i}
                className="flex items-center justify-between p-2.5 rounded-xl border border-subtle bg-surface/50"
              >
                <div className="flex items-center gap-2.5">
                  <Skeleton variant="circle" className="w-4 h-4" />
                  <Skeleton variant="text" className="w-32 h-4" />
                </div>
                <Skeleton variant="text" className="w-16 h-4" />
              </div>
            ))}
          </div>
        )}

        {/* State 2: Error */}
        {!isLoading && isError && (
          <div className="py-2">
            <ErrorBanner
              title="Ошибка загрузки тегов"
              message={error?.message || 'Не удалось получить список тегов'}
              onRetry={() => refetch()}
            />
          </div>
        )}

        {/* State 3: Empty State */}
        {!isLoading && !isError && filteredTags.length === 0 && (
          <div className="py-8">
            <EmptyState
              title={search.trim() ? 'Теги не найдены' : 'В воркспейсе нет тегов'}
              description={
                search.trim()
                  ? `По запросу «${search}» ничего не найдено`
                  : 'Создавайте теги при редактировании задач в шторке'
              }
              icon={<TagIcon className="w-8 h-8 text-muted opacity-60" />}
            />
          </div>
        )}

        {/* State 4: Success - Tag List */}
        {!isLoading && !isError && filteredTags.length > 0 && (
          <div className="max-h-[380px] overflow-y-auto space-y-1 pr-1">
            {filteredTags.map((tag) => {
              const isEditing = editingTagId === tag.id;
              const isPickerOpen = activeColorPickerTagId === tag.id;

              return (
                <div
                  key={tag.id}
                  className="group flex items-center justify-between gap-3 p-2.5 rounded-xl border border-subtle/70 bg-surface/60 hover:bg-surface hover:border-subtle transition-all duration-150"
                >
                  {/* Left: Color Picker & Title */}
                  <div className="flex items-center gap-2.5 min-w-0 flex-1 relative">
                    {/* Color dot trigger */}
                    <div className="relative shrink-0">
                      <button
                        type="button"
                        onClick={() =>
                          setActiveColorPickerTagId(isPickerOpen ? null : tag.id)
                        }
                        style={{ backgroundColor: tag.colorHex }}
                        className="w-4 h-4 rounded-full border border-subtle shadow-washi-sm hover:scale-125 transition-transform cursor-pointer"
                        title="Сменить цвет метки"
                        aria-label={`Сменить цвет тега ${tag.name}`}
                      />

                      {/* 12-Color Palette Popover */}
                      {isPickerOpen && (
                        <div
                          ref={colorPickerRef}
                          className="absolute left-0 top-full mt-2 z-50 bg-surface border border-subtle rounded-xl shadow-washi-modal p-2 w-48 text-xs space-y-1.5 animate-in fade-in zoom-in-95 duration-100"
                          onClick={(e) => e.stopPropagation()}
                        >
                          <div className="text-[10px] font-mono uppercase tracking-wider text-muted px-1">
                            Цвет метки
                          </div>
                          <div className="grid grid-cols-6 gap-1.5 p-1">
                            {TAG_PALETTE.map((c) => {
                              const isSelected =
                                (tag.colorHex || '').toLowerCase() === c.hex.toLowerCase();
                              return (
                                <button
                                  key={c.hex}
                                  type="button"
                                  title={c.name}
                                  onClick={() => handleColorSelect(tag.id, c.hex)}
                                  style={{ backgroundColor: c.hex }}
                                  className={cn(
                                    'w-5 h-5 rounded-full flex items-center justify-center transition-all hover:scale-115 cursor-pointer',
                                    isSelected
                                      ? 'ring-2 ring-primary ring-offset-2 ring-offset-surface scale-105'
                                      : 'opacity-85 hover:opacity-100'
                                  )}
                                >
                                  {isSelected && (
                                    <Check className="w-3 h-3 text-white drop-shadow-xs" />
                                  )}
                                </button>
                              );
                            })}
                          </div>
                        </div>
                      )}
                    </div>

                    {/* Tag Name (Inline Editable) */}
                    {isEditing ? (
                      <input
                        ref={editInputRef}
                        type="text"
                        value={editingName}
                        onChange={(e) => setEditingName(e.target.value)}
                        onBlur={() => handleSaveRename(tag.id)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') {
                            e.preventDefault();
                            handleSaveRename(tag.id);
                          } else if (e.key === 'Escape') {
                            e.preventDefault();
                            setEditingTagId(null);
                          }
                        }}
                        className="bg-canvas border border-accent-solid rounded px-2 py-0.5 text-xs text-primary font-medium focus:outline-none min-w-0 flex-1"
                      />
                    ) : (
                      <span
                        onDoubleClick={() => handleStartRename(tag)}
                        className="text-xs font-medium text-primary truncate cursor-pointer hover:text-accent-solid transition-colors"
                        title="Двойной клик для переименования"
                      >
                        {tag.name}
                      </span>
                    )}
                  </div>

                  {/* Right: Task Count Badge & Actions */}
                  <div className="flex items-center gap-2 shrink-0">
                    <Badge
                      variant={tag.taskCount > 0 ? 'subtle' : 'subtle'}
                      size="sm"
                      className={cn(
                        'font-mono text-[10px]',
                        tag.taskCount === 0 && 'text-muted/60 opacity-80'
                      )}
                    >
                      {formatTaskCount(tag.taskCount)}
                    </Badge>

                    {!isEditing && (
                      <button
                        type="button"
                        onClick={() => handleStartRename(tag)}
                        className="p-1 rounded text-muted hover:text-primary transition-colors opacity-0 group-hover:opacity-100"
                        title="Переименовать"
                        aria-label={`Переименовать тег ${tag.name}`}
                      >
                        <Pencil className="w-3 h-3" />
                      </button>
                    )}

                    <button
                      type="button"
                      disabled={deleteTagMutation.isPending}
                      onClick={() => handleDeleteTag(tag)}
                      className="p-1 rounded text-muted hover:text-destructive transition-colors opacity-0 group-hover:opacity-100 disabled:opacity-30"
                      title="Удалить тег"
                      aria-label={`Удалить тег ${tag.name}`}
                    >
                      <Trash2 className="w-3 h-3" />
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </Modal>
  );
};
