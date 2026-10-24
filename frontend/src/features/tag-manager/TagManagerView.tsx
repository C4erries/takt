import React, { useState, useRef, useEffect } from 'react';
import { Tag as TagIcon, Trash2, Pencil, Check, X, Search } from 'lucide-react';
import {
  Input,
  Button,
  Badge,
  Skeleton,
  EmptyState,
  ErrorBanner,
} from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import { TAG_PALETTE, DEFAULT_TAG_COLOR } from '@/entities/task';
import type { TagWithMetrics } from '@/shared/api';
import {
  useWorkspaceTagsMetricsQuery,
  useUpdateWorkspaceTagMutation,
  useDeleteWorkspaceTagMutation,
  useDeleteUnusedTagsMutation,
} from './api';

export interface TagManagerViewProps {
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

export const TagManagerView: React.FC<TagManagerViewProps> = ({ workspaceId }) => {
  const [search, setSearch] = useState('');
  const [editingTagId, setEditingTagId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');
  const editInputRef = useRef<HTMLInputElement>(null);

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
    <div className="flex-1 flex flex-col h-full overflow-hidden p-6 md:p-8 space-y-5">
      {/* Header */}
      <div className="flex items-center justify-between gap-4 flex-wrap pb-4 border-b border-subtle">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="font-serif text-2xl md:text-3xl font-normal text-primary">
              Менеджер тегов
            </h1>
            <Badge variant="subtle" size="sm" dot>
              {tags.length} тегов
            </Badge>
          </div>
          <p className="text-xs text-muted mt-1">
            Каталог меток воркспейса, палитра цветов и очистка неиспользуемых тегов
          </p>
        </div>

        {/* Toolbar: Search & Delete Unused */}
        <div className="flex items-center gap-2.5 flex-wrap w-full sm:w-auto">
          <div className="relative flex-1 sm:w-64">
            <Search className="w-3.5 h-3.5 text-muted absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none" />
            <Input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Поиск по тегам..."
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
            Удалить неиспользуемые
            {unusedCount > 0 && (
              <Badge variant="subtle" size="sm" className="ml-1.5 text-[10px]">
                {unusedCount}
              </Badge>
            )}
          </Button>
        </div>
      </div>

      {/* State 1: Loading */}
      {isLoading && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3 flex-1 overflow-y-auto pr-1">
          {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
            <div
              key={i}
              className="p-3.5 rounded-xl border border-subtle bg-surface/50 space-y-3"
            >
              <div className="flex items-center justify-between">
                <Skeleton variant="rect" className="w-24 h-5 rounded-md" />
                <Skeleton variant="circle" className="w-5 h-5 rounded" />
              </div>
              <Skeleton variant="text" className="w-16 h-3" />
              <div className="flex items-center gap-1.5 pt-2 border-t border-subtle/50">
                {[1, 2, 3, 4, 5, 6].map((d) => (
                  <Skeleton key={d} variant="circle" className="w-4 h-4 rounded-full" />
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* State 2: Error */}
      {!isLoading && isError && (
        <div className="py-4">
          <ErrorBanner
            title="Ошибка загрузки тегов"
            message={error?.message || 'Не удалось получить список тегов воркспейса'}
            onRetry={() => refetch()}
          />
        </div>
      )}

      {/* State 3: Empty State */}
      {!isLoading && !isError && filteredTags.length === 0 && (
        <div className="flex-1 flex items-center justify-center py-12">
          <EmptyState
            title={search.trim() ? 'Теги не найдены' : 'В воркспейсе нет тегов'}
            description={
              search.trim()
                ? `По запросу «${search}» ничего не найдено`
                : 'Создавайте теги при редактировании задач в шторке задач'
            }
            icon={<TagIcon className="w-8 h-8 text-muted opacity-60" />}
          />
        </div>
      )}

      {/* State 4: Success Grid */}
      {!isLoading && !isError && filteredTags.length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3 flex-1 overflow-y-auto pr-1 pb-4">
          {filteredTags.map((tag) => {
            const isEditing = editingTagId === tag.id;
            const activeHex = (tag.colorHex || DEFAULT_TAG_COLOR).toLowerCase();

            return (
              <div
                key={tag.id}
                className="group p-3.5 rounded-xl border border-subtle/70 bg-surface/60 hover:bg-surface hover:border-subtle transition-all duration-150 flex flex-col justify-between gap-3 shadow-washi-sm"
              >
                {/* Top: Tag Title / Rename Form & Actions */}
                <div className="flex items-center justify-between gap-2">
                  {isEditing ? (
                    <div className="flex items-center gap-1 flex-1">
                      <input
                        ref={editInputRef}
                        type="text"
                        value={editingName}
                        onChange={(e) => setEditingName(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') handleSaveRename(tag.id);
                          if (e.key === 'Escape') setEditingTagId(null);
                        }}
                        className="w-full text-xs bg-canvas border border-accent-solid rounded px-2 py-1 text-primary focus:outline-none"
                      />
                      <button
                        type="button"
                        onClick={() => handleSaveRename(tag.id)}
                        className="p-1 text-accent-solid hover:bg-canvas rounded transition-colors"
                        title="Сохранить"
                      >
                        <Check className="w-3.5 h-3.5" />
                      </button>
                      <button
                        type="button"
                        onClick={() => setEditingTagId(null)}
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
                          onClick={() => handleStartRename(tag)}
                          className="p-1 rounded text-muted hover:text-primary hover:bg-canvas transition-colors"
                          title="Переименовать тег"
                        >
                          <Pencil className="w-3 h-3" />
                        </button>
                        <button
                          type="button"
                          onClick={() => handleDeleteTag(tag)}
                          className="p-1 rounded text-muted hover:text-destructive hover:bg-canvas transition-colors"
                          title="Удалить тег"
                        >
                          <Trash2 className="w-3 h-3" />
                        </button>
                      </div>
                    </>
                  )}
                </div>

                {/* Middle: Usage Metrics */}
                <div className="flex items-center justify-between text-[11px] text-muted font-mono">
                  <span>Использование:</span>
                  <Badge
                    variant="subtle"
                    size="sm"
                    className={cn('text-[10px]', tag.taskCount === 0 && 'opacity-60')}
                  >
                    {formatTaskCount(tag.taskCount)}
                  </Badge>
                </div>

                {/* Bottom: 12-Color 1-Click Palette */}
                <div className="pt-2 border-t border-subtle/50 space-y-1">
                  <span className="text-[10px] font-mono text-muted/70 block">Цвет метки:</span>
                  <div className="flex items-center gap-1.5 flex-wrap">
                    {TAG_PALETTE.map((c) => {
                      const isSelected = activeHex === c.hex.toLowerCase();
                      return (
                        <button
                          key={c.hex}
                          type="button"
                          onClick={() => handleColorSelect(tag.id, c.hex)}
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
          })}
        </div>
      )}
    </div>
  );
};
