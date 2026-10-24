import React, { useState } from 'react';
import { Archive, FolderClosed, CheckCircle2, Clock } from 'lucide-react';
import {
  Badge,
  Skeleton,
  ErrorBanner,
  EmptyState,
} from '@/shared/components/ui';
import { useProjectsQuery } from '@/entities/project';
import { useArchivedTasksQuery } from './api';
import { RestoreTaskPopover } from './RestoreTaskPopover';

export interface ArchiveViewProps {
  workspaceId?: string;
}

export const ArchiveView: React.FC<ArchiveViewProps> = ({ workspaceId }) => {
  const [selectedProjectId, setSelectedProjectId] = useState<string>('all');
  const { data: projects = [] } = useProjectsQuery(workspaceId);

  const effectiveProjectId = selectedProjectId === 'all' ? undefined : selectedProjectId;
  const {
    data: tasks = [],
    isLoading,
    isError,
    error,
    refetch,
  } = useArchivedTasksQuery(workspaceId, effectiveProjectId);

  const formatArchivedDate = (dateStr?: string | null) => {
    if (!dateStr) return 'Недавно';
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return dateStr;
    return date.toLocaleDateString('ru-RU', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="flex-1 flex flex-col h-full overflow-hidden p-6 md:p-8 space-y-5">
      {/* Header */}
      <div className="flex items-center justify-between gap-4 flex-wrap pb-4 border-b border-subtle">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="font-serif text-2xl md:text-3xl font-normal text-primary">
              Архив задач
            </h1>
            <Badge variant="subtle" size="sm" dot>
              {tasks.length} в архиве
            </Badge>
          </div>
          <p className="text-xs text-muted mt-1">
            Завершённые задачи, сгруженные из колонок досок для сохранения истории без визуального шума
          </p>
        </div>

        {/* Project Filter */}
        {projects.length > 0 && (
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-[11px] font-mono text-muted uppercase">Проект:</span>
            <div className="flex items-center gap-1.5 p-0.5 rounded-lg bg-surface border border-subtle">
              <button
                type="button"
                onClick={() => setSelectedProjectId('all')}
                className={`px-2.5 py-1 rounded-md text-xs font-medium transition-colors ${
                  selectedProjectId === 'all'
                    ? 'bg-accent-surface text-primary shadow-washi-sm'
                    : 'text-muted hover:text-primary'
                }`}
              >
                Все проекты
              </button>
              {projects.map((p) => (
                <button
                  key={p.id}
                  type="button"
                  onClick={() => setSelectedProjectId(p.id)}
                  className={`px-2.5 py-1 rounded-md text-xs font-medium transition-colors ${
                    selectedProjectId === p.id
                      ? 'bg-accent-surface text-primary shadow-washi-sm'
                      : 'text-muted hover:text-primary'
                  }`}
                >
                  {p.name}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* State 1: Loading */}
      {isLoading && (
        <div className="space-y-3 flex-1 overflow-y-auto">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className="p-4 rounded-xl border border-subtle bg-surface/50 flex items-center justify-between gap-4"
            >
              <div className="space-y-2 flex-1">
                <Skeleton variant="text" className="w-1/3 h-4" />
                <Skeleton variant="text" className="w-1/4 h-3" />
              </div>
              <Skeleton variant="rect" className="w-28 h-8 rounded-lg" />
            </div>
          ))}
        </div>
      )}

      {/* State 2: Error */}
      {!isLoading && isError && (
        <div className="py-6">
          <ErrorBanner
            title="Ошибка загрузки архива задач"
            message={error?.message || 'Не удалось получить данные с сервера'}
            onRetry={() => refetch()}
          />
        </div>
      )}

      {/* State 3: Empty State */}
      {!isLoading && !isError && tasks.length === 0 && (
        <div className="flex-1 flex items-center justify-center py-16">
          <EmptyState
            title="Архив пуст"
            description="Здесь будут храниться завершённые задачи, сгруженные из колонок досок через меню «Сгрузить выполненные в архив»."
            icon={<Archive className="w-8 h-8 text-muted opacity-60" />}
          />
        </div>
      )}

      {/* State 4: Success List */}
      {!isLoading && !isError && tasks.length > 0 && (
        <div className="space-y-2.5 flex-1 overflow-y-auto pr-1">
          {tasks.map((task) => {
            const project = projects.find((p) => p.id === task.projectId);
            return (
              <div
                key={task.id}
                className="p-3.5 rounded-xl border border-subtle bg-surface/70 hover:border-subtle/90 transition-all shadow-washi-sm flex items-center justify-between gap-4 flex-wrap sm:flex-nowrap"
              >
                <div className="space-y-1.5 min-w-0 flex-1">
                  <div className="flex items-center gap-2 flex-wrap">
                    <CheckCircle2 className="w-4 h-4 text-accent-solid shrink-0" />
                    <span className="font-serif text-sm font-normal text-primary break-words">
                      {task.title}
                    </span>
                    {project && (
                      <Badge variant="subtle" size="sm" className="font-sans text-[10px]">
                        <FolderClosed className="w-3 h-3 mr-1 inline-block" />
                        {project.name}
                      </Badge>
                    )}
                  </div>

                  <div className="flex items-center gap-2.5 flex-wrap text-[11px] font-mono text-muted">
                    <span className="flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      <span>Сгружено: {formatArchivedDate(task.archivedAt ?? task.updatedAt)}</span>
                    </span>

                    {task.tags && task.tags.length > 0 && (
                      <div className="flex items-center gap-1">
                        {task.tags.map((t) => (
                          <span
                            key={t.id}
                            className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-full bg-canvas/60 border border-subtle text-[10px]"
                          >
                            {t.colorHex && (
                              <span
                                className="w-1.5 h-1.5 rounded-full shrink-0"
                                style={{ backgroundColor: t.colorHex }}
                              />
                            )}
                            <span>{t.name}</span>
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>

                <div className="shrink-0 ml-auto sm:ml-0">
                  <RestoreTaskPopover
                    task={task}
                    workspaceId={workspaceId}
                    onRestored={() => refetch()}
                  />
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
