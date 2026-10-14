import { useState, useEffect, useCallback } from 'react';
import type { Task } from '@/shared/api';

export interface UseTaskNavigationProps {
  tasks: Task[];
  onToggleTask?: (taskId: string) => void;
  isEnabled?: boolean;
}

export interface UseTaskNavigationReturn {
  selectedIndex: number;
  setSelectedIndex: (index: number) => void;
  selectedTaskId: string | null;
}

export const useTaskNavigation = ({
  tasks,
  onToggleTask,
  isEnabled = true,
}: UseTaskNavigationProps): UseTaskNavigationReturn => {
  const [selectedIndex, setSelectedIndex] = useState<number>(0);

  // Keep index within bounds if tasks list changes
  useEffect(() => {
    if (tasks.length === 0) {
      setSelectedIndex(0);
    } else if (selectedIndex >= tasks.length) {
      setSelectedIndex(tasks.length - 1);
    }
  }, [tasks.length, selectedIndex]);

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
        setSelectedIndex((prev) => Math.min(prev + 1, tasks.length - 1));
      } else if (key === 'k' || key === 'K' || key === 'ArrowUp') {
        e.preventDefault();
        setSelectedIndex((prev) => Math.max(prev - 1, 0));
      } else if (key === ' ' || key === 'x' || key === 'X') {
        if (onToggleTask && tasks[selectedIndex]) {
          e.preventDefault();
          onToggleTask(tasks[selectedIndex].id);
        }
      }
    },
    [isEnabled, tasks, selectedIndex, onToggleTask]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [handleKeyDown]);

  const selectedTaskId = tasks[selectedIndex]?.id ?? null;

  return {
    selectedIndex,
    setSelectedIndex,
    selectedTaskId,
  };
};
