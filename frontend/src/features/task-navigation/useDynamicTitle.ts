import { useEffect } from 'react';
import { useTaskStore } from './taskStore';
import { useProjectBoardsQuery } from '@/entities/project';

export function useDynamicTitle(workspaceId?: string): void {
  const { activeNav, selectedProjectId, activeBoardId } = useTaskStore();
  const boardsQuery = useProjectBoardsQuery(workspaceId, selectedProjectId ?? undefined);
  const boards = boardsQuery.data ?? [];

  useEffect(() => {
    if (typeof document === 'undefined') return;

    if (activeNav === 'inbox') {
      document.title = 'Входящие | Takt';
    } else if (activeNav === 'today') {
      document.title = 'Сегодня | Takt';
    } else if (activeNav === 'automations') {
      document.title = 'Автоматизации | Takt';
    } else if (activeNav === 'archive') {
      document.title = 'Архив | Takt';
    } else if (activeNav === 'tags') {
      document.title = 'Теги | Takt';
    } else if (activeNav === 'projects') {
      const activeBoard = boards.find((b) => b.id === activeBoardId);
      document.title = activeBoard ? `${activeBoard.name} | Takt` : 'Takt';
    } else {
      document.title = 'Takt';
    }
  }, [activeNav, activeBoardId, boards]);
}
