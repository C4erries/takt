import React, { useState, useRef, useEffect } from 'react';
import { RotateCcw, Check, ChevronDown } from 'lucide-react';
import { Button } from '@/shared/components/ui';
import type { Task } from '@/entities/task';
import {
  useProjectsQuery,
  useProjectBoardsQuery,
  useProjectAllColumnsQuery,
} from '@/entities/project';
import { useUnarchiveTaskMutation } from './api';

export interface RestoreTaskPopoverProps {
  task: Task;
  workspaceId?: string;
  onRestored?: () => void;
}

export const RestoreTaskPopover: React.FC<RestoreTaskPopoverProps> = ({
  task,
  workspaceId,
  onRestored,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const { data: projects = [] } = useProjectsQuery(workspaceId);
  const [selectedProjectId, setSelectedProjectId] = useState<string>(
    task.projectId || projects[0]?.id || ''
  );

  useEffect(() => {
    if (!selectedProjectId && projects.length > 0) {
      setSelectedProjectId(task.projectId || projects[0].id);
    }
  }, [projects, selectedProjectId, task.projectId]);

  const { data: boards = [] } = useProjectBoardsQuery(workspaceId, selectedProjectId || undefined);
  const { data: boardGroups = [] } = useProjectAllColumnsQuery(workspaceId, boards);

  const [selectedBoardId, setSelectedBoardId] = useState<string>('');
  const [selectedColumnId, setSelectedColumnId] = useState<string>('');

  useEffect(() => {
    if (boardGroups.length > 0) {
      const activeGroup = boardGroups.find((bg) => bg.board.id === selectedBoardId) || boardGroups[0];
      setSelectedBoardId(activeGroup.board.id);
      if (activeGroup.columns.length > 0 && !activeGroup.columns.some((c) => c.id === selectedColumnId)) {
        setSelectedColumnId(activeGroup.columns[0].id);
      }
    }
  }, [boardGroups, selectedBoardId, selectedColumnId]);

  const unarchiveMutation = useUnarchiveTaskMutation(workspaceId);

  useEffect(() => {
    if (!isOpen) return;
    const onOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', onOutside);
    return () => document.removeEventListener('mousedown', onOutside);
  }, [isOpen]);

  const handleRestore = async () => {
    if (!selectedColumnId) return;
    try {
      await unarchiveMutation.mutateAsync({
        taskId: task.id,
        data: {
          columnId: selectedColumnId,
          boardId: selectedBoardId || undefined,
        },
      });
      setIsOpen(false);
      onRestored?.();
    } catch {
      // Handled in mutation
    }
  };

  const currentGroup = boardGroups.find((bg) => bg.board.id === selectedBoardId);
  const availableColumns = currentGroup?.columns ?? [];

  return (
    <div className="relative inline-block" ref={containerRef}>
      <Button
        variant="outline"
        size="sm"
        onClick={() => setIsOpen((prev) => !prev)}
        leftIcon={<RotateCcw className="w-3.5 h-3.5 text-accent-solid" />}
        className="text-xs h-8 bg-surface/80 shadow-washi-sm"
      >
        <span>Восстановить</span>
        <ChevronDown className="w-3 h-3 text-muted ml-1" />
      </Button>

      {isOpen && (
        <div className="absolute right-0 top-full mt-1.5 w-64 p-3 bg-surface border border-subtle rounded-xl shadow-washi-modal z-30 text-xs space-y-3 animate-in fade-in zoom-in-95 duration-100">
          <div className="space-y-0.5">
            <h4 className="font-medium text-primary">Восстановление задачи</h4>
            <p className="text-[11px] text-muted leading-tight">
              Выберите доску и колонку для возврата карточки
            </p>
          </div>

          {projects.length > 1 && (
            <div className="space-y-1">
              <label className="text-[10px] font-mono uppercase tracking-wider text-muted">
                Проект
              </label>
              <select
                value={selectedProjectId}
                onChange={(e) => {
                  setSelectedProjectId(e.target.value);
                  setSelectedBoardId('');
                  setSelectedColumnId('');
                }}
                className="w-full text-xs bg-canvas border border-subtle rounded-lg px-2 py-1.5 text-primary focus:outline-none focus:border-accent-solid"
              >
                {projects.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {boardGroups.length > 0 && (
            <div className="space-y-1">
              <label className="text-[10px] font-mono uppercase tracking-wider text-muted">
                Доска
              </label>
              <select
                value={selectedBoardId}
                onChange={(e) => {
                  setSelectedBoardId(e.target.value);
                  const grp = boardGroups.find((bg) => bg.board.id === e.target.value);
                  if (grp && grp.columns.length > 0) setSelectedColumnId(grp.columns[0].id);
                }}
                className="w-full text-xs bg-canvas border border-subtle rounded-lg px-2 py-1.5 text-primary focus:outline-none focus:border-accent-solid"
              >
                {boardGroups.map((bg) => (
                  <option key={bg.board.id} value={bg.board.id}>
                    {bg.boardName}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="space-y-1">
            <label className="text-[10px] font-mono uppercase tracking-wider text-muted">
              Колонка
            </label>
            <select
              value={selectedColumnId}
              onChange={(e) => setSelectedColumnId(e.target.value)}
              disabled={availableColumns.length === 0}
              className="w-full text-xs bg-canvas border border-subtle rounded-lg px-2 py-1.5 text-primary focus:outline-none focus:border-accent-solid disabled:opacity-50"
            >
              {availableColumns.length === 0 && <option value="">Нет колонок</option>}
              {availableColumns.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>

          <div className="flex items-center justify-end gap-1.5 pt-2 border-t border-subtle/50">
            <Button variant="ghost" size="sm" onClick={() => setIsOpen(false)} className="h-7 text-xs px-2">
              Отмена
            </Button>
            <Button
              variant="solid"
              size="sm"
              onClick={handleRestore}
              disabled={!selectedColumnId || unarchiveMutation.isPending}
              isLoading={unarchiveMutation.isPending}
              leftIcon={<Check className="w-3.5 h-3.5" />}
              className="h-7 text-xs px-2.5"
            >
              Вернуть
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};
