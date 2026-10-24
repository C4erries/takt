import React, { useState, useEffect, useRef } from 'react';
import { Link2, Search, X, Loader2 } from 'lucide-react';
import { Input, Skeleton } from '@/shared/components/ui';
import { searchWorkspace, type Task } from '@/shared/api';
import { formatTaskNumber } from '@/entities/task';
import { cn } from '@/shared/lib/cn';

export interface AttachTaskPopoverProps {
  workspaceId?: string;
  currentTaskId: string;
  onAttach: (taskId: string) => Promise<void>;
  onClose: () => void;
  className?: string;
}

export const AttachTaskPopover: React.FC<AttachTaskPopoverProps> = ({
  workspaceId,
  currentTaskId,
  onAttach,
  onClose,
  className,
}) => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Task[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [attachingId, setAttachingId] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const onMouseDown = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        onClose();
      }
    };
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('mousedown', onMouseDown);
    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('mousedown', onMouseDown);
      document.removeEventListener('keydown', onKeyDown);
    };
  }, [onClose]);

  useEffect(() => {
    const trimmed = query.trim();
    if (!trimmed || !workspaceId) {
      setResults([]);
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    const timer = setTimeout(async () => {
      try {
        const res = await searchWorkspace(workspaceId, trimmed);
        const filtered = (res.tasks || []).filter((t) => {
          if (t.id === currentTaskId) return false;
          if (t.parentId) return false;
          if ((t.subtasksCount ?? 0) > 0) return false;
          if (t.subtasks && t.subtasks.length > 0) return false;
          return true;
        });
        setResults(filtered);
      } catch {
        setResults([]);
      } finally {
        setIsLoading(false);
      }
    }, 200);

    return () => clearTimeout(timer);
  }, [query, workspaceId, currentTaskId]);

  const handleSelect = async (taskId: string) => {
    if (attachingId) return;
    setAttachingId(taskId);
    try {
      await onAttach(taskId);
      onClose();
    } finally {
      setAttachingId(null);
    }
  };

  return (
    <div
      ref={containerRef}
      className={cn(
        'absolute left-0 right-0 top-full mt-1.5 z-50 bg-surface border border-subtle rounded-xl shadow-washi-modal p-2.5 text-xs space-y-2 animate-in fade-in zoom-in-95 duration-100 min-w-[280px]',
        className
      )}
      onClick={(e) => e.stopPropagation()}
    >
      <div className="flex items-center justify-between pb-1.5 border-b border-subtle/50">
        <div className="flex items-center gap-1.5 text-primary font-medium text-xs">
          <Link2 className="w-3.5 h-3.5 text-accent-solid" />
          <span>Привязать существующую задачу</span>
        </div>
        <button
          type="button"
          onClick={onClose}
          className="text-muted hover:text-primary transition-colors p-0.5 rounded hover:bg-canvas/50"
          title="Закрыть"
        >
          <X className="w-3.5 h-3.5" />
        </button>
      </div>

      <div className="relative">
        <Search className="w-3.5 h-3.5 text-muted absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none" />
        <Input
          autoFocus
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Поиск задачи по названию..."
          className="text-xs bg-surface border-subtle h-8 w-full pl-8"
        />
      </div>

      <div className="max-h-48 overflow-y-auto space-y-1 pt-0.5">
        {isLoading && (
          <div className="space-y-1.5 p-1">
            <Skeleton variant="text" className="w-full h-5" />
            <Skeleton variant="text" className="w-4/5 h-5" />
          </div>
        )}

        {!isLoading && !query.trim() && (
          <p className="px-2 py-3 text-[11px] text-muted italic text-center">
            Введите название задачи для поиска...
          </p>
        )}

        {!isLoading && query.trim() && results.length === 0 && (
          <p className="px-2 py-3 text-[11px] text-muted italic text-center">
            Подходящих задач не найдено
          </p>
        )}

        {!isLoading &&
          results.map((task) => (
            <button
              key={task.id}
              type="button"
              disabled={Boolean(attachingId)}
              onClick={() => handleSelect(task.id)}
              className="w-full text-left px-2.5 py-1.5 rounded-lg hover:bg-canvas/70 text-primary transition-colors flex items-center justify-between group disabled:opacity-50"
            >
              <div className="flex items-center gap-2 min-w-0 flex-1 pr-2">
                <span className="font-medium truncate text-xs">{task.title}</span>
                <span className="font-mono text-[10px] text-muted shrink-0">
                  {formatTaskNumber(task.id)}
                </span>
              </div>
              <div className="shrink-0 flex items-center">
                {attachingId === task.id ? (
                  <Loader2 className="w-3 h-3 text-accent-solid animate-spin" />
                ) : (
                  <span className="text-[10px] text-accent-solid opacity-0 group-hover:opacity-100 transition-opacity font-mono">
                    Привязать ↵
                  </span>
                )}
              </div>
            </button>
          ))}
      </div>
    </div>
  );
};
