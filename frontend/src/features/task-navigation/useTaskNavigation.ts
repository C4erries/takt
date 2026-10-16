import { useEffect, useCallback, useState } from 'react';
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
  const [cursorIndex, setCursorIndex] = useState<number>(0);

  // Keep cursorIndex valid when tasks list changes
  useEffect(() => {
    if (cursorIndex >= tasks.length && tasks.length > 0) {
      setCursorIndex(tasks.length - 1);
    }
  }, [tasks.length, cursorIndex]);

  // Sync cursorIndex with selectedTaskId if selectedTaskId exists in tasks
  useEffect(() => {
    if (selectedTaskId) {
      const idx = tasks.findIndex((t) => t.id === selectedTaskId);
      if (idx >= 0) {
        setCursorIndex(idx);
      }
    }
  }, [selectedTaskId, tasks]);

  const setSelectedIndex = useCallback(
    (index: number) => {
      setCursorIndex(index);
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
        const nextIndex = Math.min(cursorIndex + 1, tasks.length - 1);
        setCursorIndex(nextIndex);
        if (selectedTaskId && tasks[nextIndex]) {
          setSelectedTaskId(tasks[nextIndex].id);
        }
      } else if (key === 'k' || key === 'K' || key === 'ArrowUp') {
        e.preventDefault();
        const prevIndex = Math.max(cursorIndex - 1, 0);
        setCursorIndex(prevIndex);
        if (selectedTaskId && tasks[prevIndex]) {
          setSelectedTaskId(tasks[prevIndex].id);
        }
      } else if (key === 'Enter') {
        if (tasks[cursorIndex]) {
          e.preventDefault();
          setSelectedTaskId(tasks[cursorIndex].id);
        }
      } else if (key === 'Escape') {
        if (selectedTaskId) {
          e.preventDefault();
          setSelectedTaskId(null);
        }
      } else if (key === ' ' || key === 'x' || key === 'X') {
        if (onToggleTask && tasks[cursorIndex]) {
          e.preventDefault();
          onToggleTask(tasks[cursorIndex].id);
        }
      }
    },
    [isEnabled, tasks, cursorIndex, selectedTaskId, setSelectedTaskId, onToggleTask]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [handleKeyDown]);

  return {
    selectedIndex: cursorIndex,
    setSelectedIndex,
    selectedTaskId,
    setSelectedTaskId,
  };
};
