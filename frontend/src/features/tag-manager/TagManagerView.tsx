import React, { useState, useRef, useEffect, useMemo } from 'react';
import { Tag as TagIcon, Trash2, Search } from 'lucide-react';
import {
  Input,
  Button,
  Badge,
  Skeleton,
  EmptyState,
  ErrorBanner,
} from '@/shared/components/ui';
import { useCreateTagMutation } from '@/entities/task';
import { useProjectsQuery } from '@/entities/project';
import type { TagWithMetrics } from '@/shared/api';
import {
  useWorkspaceTagsMetricsQuery,
  useUpdateWorkspaceTagMutation,
  useDeleteWorkspaceTagMutation,
  useDeleteUnusedTagsMutation,
} from './api';
import { TagCard } from './TagCard';
import { CreateTagBar } from './CreateTagBar';

export interface TagManagerViewProps {
  workspaceId?: string;
}

export type ScopeFilter = 'all' | 'global' | string;

export const TagManagerView: React.FC<TagManagerViewProps> = ({ workspaceId }) => {
  const [search, setSearch] = useState('');
  const [scopeFilter, setScopeFilter] = useState<ScopeFilter>('all');
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

  const { data: projects = [] } = useProjectsQuery(workspaceId);

  const createTagMutation = useCreateTagMutation(workspaceId);
  const updateTagMutation = useUpdateWorkspaceTagMutation(workspaceId);
  const deleteTagMutation = useDeleteWorkspaceTagMutation(workspaceId);
  const deleteUnusedMutation = useDeleteUnusedTagsMutation(workspaceId);

  useEffect(() => {
    if (editingTagId) {
      editInputRef.current?.focus();
      editInputRef.current?.select();
    }
  }, [editingTagId]);

  const projectsMap = useMemo(() => {
    const map = new Map<string, string>();
    projects.forEach((p) => map.set(p.id, p.name));
    return map;
  }, [projects]);

  const unusedTags = tags.filter((t) => t.taskCount === 0);
  const unusedCount = unusedTags.length;

  const filteredTags = useMemo(() => {
    return tags.filter((t) => {
      // Scope filter
      if (scopeFilter === 'global') {
        if (t.projectId) return false;
      } else if (scopeFilter !== 'all') {
        // specific project id
        if (t.projectId !== scopeFilter) return false;
      }

      // Search filter
      if (search.trim()) {
        const matchesName = t.name.toLowerCase().includes(search.trim().toLowerCase());
        const projName = t.projectId ? projectsMap.get(t.projectId) || '' : '';
        const matchesProject = projName.toLowerCase().includes(search.trim().toLowerCase());
        return matchesName || matchesProject;
      }
      return true;
    });
  }, [tags, scopeFilter, search, projectsMap]);

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
        ? `Тег «${tag.name}» будет отвязан от ${tag.taskCount} задач. Продолжить?`
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

  const handleCreateTag = async (data: { name: string; colorHex: string; projectId?: string | null }) => {
    await createTagMutation.mutateAsync({
      name: data.name,
      colorHex: data.colorHex,
      projectId: data.projectId,
    });
  };

  return (
    <div className="flex-1 flex flex-col h-full overflow-hidden p-6 md:p-8 space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between gap-4 flex-wrap pb-3 border-b border-subtle">
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
            Каталог меток, разграничение по проектам и очистка неиспользуемых тегов
          </p>
        </div>

        <div className="flex items-center gap-2 flex-wrap sm:flex-nowrap">
          <CreateTagBar
            projects={projects}
            defaultProjectId={scopeFilter !== 'all' && scopeFilter !== 'global' ? scopeFilter : null}
            isPending={createTagMutation.isPending}
            onCreateTag={handleCreateTag}
          />

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
            Очистить пустые
            {unusedCount > 0 && (
              <Badge variant="subtle" size="sm" className="ml-1.5 text-[10px]">
                {unusedCount}
              </Badge>
            )}
          </Button>
        </div>
      </div>

      {/* Toolbar: Scope Filter Chips & Search */}
      <div className="flex items-center justify-between gap-3 flex-wrap">
        {/* Scope Filter Chips */}
        <div className="flex items-center gap-1.5 flex-wrap">
          <button
            type="button"
            onClick={() => setScopeFilter('all')}
            className={`px-2.5 py-1 rounded-lg text-xs transition-colors ${
              scopeFilter === 'all'
                ? 'bg-accent-surface font-medium text-primary shadow-washi-sm'
                : 'text-muted hover:text-primary hover:bg-surface/60'
            }`}
          >
            Все ({tags.length})
          </button>

          <button
            type="button"
            onClick={() => setScopeFilter('global')}
            className={`px-2.5 py-1 rounded-lg text-xs transition-colors ${
              scopeFilter === 'global'
                ? 'bg-accent-surface font-medium text-primary shadow-washi-sm'
                : 'text-muted hover:text-primary hover:bg-surface/60'
            }`}
          >
            Глобальные ({tags.filter((t) => !t.projectId).length})
          </button>

          {projects.length > 0 && (
            <div className="flex items-center gap-1 pl-1 border-l border-subtle/70">
              <span className="text-[10px] font-mono text-muted pl-1 hidden sm:inline">Проект:</span>
              <select
                value={scopeFilter !== 'all' && scopeFilter !== 'global' ? scopeFilter : ''}
                onChange={(e) => {
                  const val = e.target.value;
                  setScopeFilter(val || 'all');
                }}
                className={`text-xs bg-surface border rounded-lg px-2 py-1 text-primary focus:outline-none focus:border-accent-solid ${
                  scopeFilter !== 'all' && scopeFilter !== 'global'
                    ? 'border-accent-solid text-primary bg-accent-surface/30'
                    : 'border-subtle text-muted'
                }`}
              >
                <option value="">Все проекты...</option>
                {projects.map((p) => {
                  const pTagsCount = tags.filter((t) => t.projectId === p.id).length;
                  return (
                    <option key={p.id} value={p.id}>
                      {p.name} ({pTagsCount})
                    </option>
                  );
                })}
              </select>
            </div>
          )}
        </div>

        {/* Search */}
        <div className="relative w-full sm:w-60">
          <Search className="w-3.5 h-3.5 text-muted absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none" />
          <Input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Поиск по тегам..."
            className="text-xs bg-surface border-subtle h-8 w-full pl-8"
          />
        </div>
      </div>

      {/* State 1: Loading */}
      {isLoading && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3 flex-1 overflow-y-auto pr-1 content-start items-start auto-rows-min">
          {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
            <div
              key={i}
              className="p-3 rounded-xl border border-subtle bg-surface/50 space-y-2 min-h-[130px] max-h-[155px]"
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
            title={search.trim() ? 'Теги не найдены' : 'Теги в этой категории отсутствуют'}
            description={
              search.trim()
                ? `По запросу «${search}» ничего не найдено`
                : 'Создайте первый тег с помощью кнопки «Создать тег» вверху'
            }
            icon={<TagIcon className="w-8 h-8 text-muted opacity-60" />}
          />
        </div>
      )}

      {/* State 4: Success Grid */}
      {!isLoading && !isError && filteredTags.length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3 flex-1 overflow-y-auto pr-1 pb-4 content-start items-start auto-rows-min">
          {filteredTags.map((tag) => (
            <TagCard
              key={tag.id}
              tag={tag}
              projectName={tag.projectId ? projectsMap.get(tag.projectId) : undefined}
              isEditing={editingTagId === tag.id}
              editingName={editingName}
              editInputRef={editInputRef}
              onStartRename={handleStartRename}
              onEditingNameChange={setEditingName}
              onSaveRename={handleSaveRename}
              onCancelRename={() => setEditingTagId(null)}
              onColorSelect={handleColorSelect}
              onDeleteTag={handleDeleteTag}
            />
          ))}
        </div>
      )}
    </div>
  );
};
