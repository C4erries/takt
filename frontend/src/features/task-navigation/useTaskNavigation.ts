import { useEffect, useCallback, useMemo } from 'react';
import type { Task } from '@/shared/api';
import { useTaskStore } from './taskStore';

export interface UseTaskNavigationProps {
  tasks: Task[];
  onToggleTask?: (taskId: string) => void;
  isEnabled?: boolean;
}

export interface UseTaskNavigationReturn {
  selectedIndex: number;
  setSelectedIndex: (index: number) => void;
  selectedTaskId: string | null;
  setSelectedTaskId: (id: string | null) => void;
}

export const useTaskNavigation = ({
  tasks,
  onToggleTask,
  isEnabled = true,
}: UseTaskNavigationProps): UseTaskNavigationReturn => {
  const { selectedTaskId, setSelectedTaskId } = useTaskStore();

  const selectedIndex = useMemo(() => {
    if (tasks.length === 0) return 0;
    if (!selectedTaskId) return 0;
    const idx = tasks.findIndex((t) => t.id === selectedTaskId);
    return idx >= 0 ? idx : 0;
  }, [tasks, selectedTaskId]);

  useEffect(() => {
    if (tasks.length > 0) {
      const exists = tasks.some((t) => t.id === selectedTaskId);
      if (!exists && tasks[0]) {
        setSelectedTaskId(tasks[0].id);
      }
    } else if (selectedTaskId !== null) {
      setSelectedTaskId(null);
    }
  }, [tasks, selectedTaskId, setSelectedTaskId]);

  const setSelectedIndex = useCallback(
    (index: number) => {
      if (tasks[index]) {
        setSelectedTaskId(tasks[index].id);
      }
    },
    [tasks, setSelectedTaskId]
  );

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (!isEnabled || tasks.length === 0) return;

      const target = e.target as HTMLElement | null;
      const isInput =
        target &&
        (target.tagName === 'INPUT' ||
          target.tagName === 'TEXTAREA' ||
          target.isContentEditable);

      if (isInput) return;
      if (e.metaKey || e.ctrlKey || e.altKey) return;

      const key = e.key;

      if (key === 'j' || key === 'J' || key === 'ArrowDown') {
        e.preventDefault();
        const nextIndex = Math.min(selectedIndex + 1, tasks.length - 1);
        if (tasks[nextIndex]) {
          setSelectedTaskId(tasks[nextIndex].id);
        }
      } else if (key === 'k' || key === 'K' || key === 'ArrowUp') {
        e.preventDefault();
        const prevIndex = Math.max(selectedIndex - 1, 0);
        if (tasks[prevIndex]) {
          setSelectedTaskId(tasks[prevIndex].id);
        }
      } else if (key === ' ' || key === 'x' || key === 'X') {
        if (onToggleTask && tasks[selectedIndex]) {
          e.preventDefault();
          onToggleTask(tasks[selectedIndex].id);
        }
      }
    },
    [isEnabled, tasks, selectedIndex, setSelectedTaskId, onToggleTask]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [handleKeyDown]);

  const currentSelectedTaskId = tasks[selectedIndex]?.id ?? selectedTaskId;

  return {
    selectedIndex,
    setSelectedIndex,
    selectedTaskId: currentSelectedTaskId,
    setSelectedTaskId,
  };
};
