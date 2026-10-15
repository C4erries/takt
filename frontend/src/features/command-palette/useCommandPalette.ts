import { useState, useEffect, useRef, useMemo } from 'react';
import { useTaskStore } from '@/features/task-navigation';
import { useProjectsQuery } from '@/entities/project';
import { searchWorkspace } from '@/shared/api';
import type { Task, Project } from '@/shared/api';
import type { CommandItem } from './types';

interface UseCommandPaletteOptions {
  isOpen: boolean;
  onClose: () => void;
  workspaceId?: string;
  onOpenQuickCapture?: () => void;
}

export const useCommandPalette = ({
  isOpen,
  onClose,
  workspaceId,
  onOpenQuickCapture,
}: UseCommandPaletteOptions) => {
  const [query, setQuery] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [searchResults, setSearchResults] = useState<{ tasks: Task[]; projects: Project[] }>({
    tasks: [],
    projects: [],
  });
  const [selectedIndex, setSelectedIndex] = useState(0);

  const { setActiveNav, setSelectedProjectId, setSelectedTaskId } = useTaskStore();
  const { data: allProjects = [] } = useProjectsQuery(workspaceId);

  const debounceTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Сброс состояния при открытии/закрытии
  useEffect(() => {
    if (isOpen) {
      setQuery('');
      setSearchResults({ tasks: [], projects: [] });
      setSelectedIndex(0);
    }
  }, [isOpen]);

  // Debounced поиск при вводе
  useEffect(() => {
    if (!isOpen) return;

    const trimmed = query.trim();
    if (!trimmed) {
      setSearchResults({ tasks: [], projects: [] });
      setIsLoading(false);
      setSelectedIndex(0);
      return;
    }

    if (debounceTimeoutRef.current) {
      clearTimeout(debounceTimeoutRef.current);
    }

    setIsLoading(true);
    debounceTimeoutRef.current = setTimeout(async () => {
      if (!workspaceId) {
        setIsLoading(false);
        return;
      }
      try {
        const response = await searchWorkspace(workspaceId, trimmed);
        setSearchResults(response);
      } catch {
        setSearchResults({ tasks: [], projects: [] });
      } finally {
        setIsLoading(false);
        setSelectedIndex(0);
      }
    }, 250);

    return () => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current);
      }
    };
  }, [query, isOpen, workspaceId]);

  // Формирование плоского списка доступных команд и элементов
  const items = useMemo<CommandItem[]>(() => {
    const list: CommandItem[] = [];
    const trimmed = query.trim();

    if (!trimmed) {
      // 1. Быстрые переходы и действия
      list.push({
        id: 'nav-inbox',
        type: 'navigation',
        title: 'Входящие (Inbox)',
        subtitle: 'Перейти к списку необработанных задач',
        badge: 'G I',
        onSelect: () => {
          setSelectedProjectId(null);
          setActiveNav('inbox');
          onClose();
        },
      });

      list.push({
        id: 'nav-today',
        type: 'navigation',
        title: 'Фокус Сегодня (Today)',
        subtitle: 'Перейти к задачам на сегодня',
        badge: 'G T',
        onSelect: () => {
          setSelectedProjectId(null);
          setActiveNav('today');
          onClose();
        },
      });

      list.push({
        id: 'action-quick-capture',
        type: 'action',
        title: 'Создать задачу...',
        subtitle: 'Быстрый захват новой задачи',
        badge: 'C',
        onSelect: () => {
          onClose();
          onOpenQuickCapture?.();
        },
      });

      // 2. Список проектов
      allProjects.forEach((proj) => {
        list.push({
          id: `project-${proj.id}`,
          type: 'project',
          title: proj.name,
          subtitle: proj.description || 'Проект воркспейса',
          color: proj.color,
          projectData: proj,
          onSelect: () => {
            setSelectedProjectId(proj.id);
            setActiveNav(`project:${proj.id}`);
            onClose();
          },
        });
      });
    } else {
      // 1. Результаты поиска по проектам
      searchResults.projects.forEach((proj) => {
        list.push({
          id: `search-project-${proj.id}`,
          type: 'project',
          title: proj.name,
          subtitle: proj.description || 'Проект',
          color: proj.color,
          projectData: proj,
          onSelect: () => {
            setSelectedProjectId(proj.id);
            setActiveNav(`project:${proj.id}`);
            onClose();
          },
        });
      });

      // 2. Результаты поиска по задачам
      searchResults.tasks.forEach((task) => {
        list.push({
          id: `search-task-${task.id}`,
          type: 'task',
          title: task.title,
          subtitle: task.descriptionMarkdown || undefined,
          isCompleted: task.isCompleted,
          taskData: task,
          onSelect: () => {
            setSelectedTaskId(task.id);
            if (task.projectId) {
              setSelectedProjectId(task.projectId);
              setActiveNav(`project:${task.projectId}`);
            }
            onClose();
          },
        });
      });
    }

    return list;
  }, [
    query,
    searchResults,
    allProjects,
    setActiveNav,
    setSelectedProjectId,
    setSelectedTaskId,
    onClose,
    onOpenQuickCapture,
  ]);

  // Клавиатурная навигация
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((prev) => (items.length > 0 ? (prev + 1) % items.length : 0));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((prev) => (items.length > 0 ? (prev - 1 + items.length) % items.length : 0));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const current = items[selectedIndex];
      if (current) {
        current.onSelect();
      }
    } else if (e.key === 'Escape') {
      e.preventDefault();
      onClose();
    }
  };

  return {
    query,
    setQuery,
    isLoading,
    items,
    selectedIndex,
    setSelectedIndex,
    handleKeyDown,
  };
};
