import { create } from 'zustand';

export type NavDestination = 'inbox' | 'today' | 'projects' | (string & {});

export interface TaskNavigationState {
  selectedTaskId: string | null;
  setSelectedTaskId: (id: string | null) => void;
  isQuickCaptureOpen: boolean;
  setQuickCaptureOpen: (isOpen: boolean) => void;
  activeFilter: string;
  setActiveFilter: (filter: string) => void;
  activeNav: NavDestination;
  setActiveNav: (nav: NavDestination) => void;
  selectedProjectId: string | null;
  setSelectedProjectId: (id: string | null) => void;
  activeBoardId: string | null;
  setActiveBoardId: (id: string | null) => void;
  isBoardsPanelCollapsed: boolean;
  setBoardsPanelCollapsed: (collapsed: boolean) => void;
  toggleBoardsPanel: () => void;
  viewMode: 'kanban' | 'list';
  setViewMode: (mode: 'kanban' | 'list') => void;
  selectedDateFilter: string | null;
  setSelectedDateFilter: (date: string | null) => void;
}

export const useTaskStore = create<TaskNavigationState>((set) => ({
  selectedTaskId: null,
  setSelectedTaskId: (selectedTaskId) => set({ selectedTaskId }),
  isQuickCaptureOpen: false,
  setQuickCaptureOpen: (isQuickCaptureOpen) => set({ isQuickCaptureOpen }),
  activeFilter: 'all',
  setActiveFilter: (activeFilter) => set({ activeFilter }),
  activeNav: 'inbox',
  setActiveNav: (activeNav) => set({ activeNav }),
  selectedProjectId: null,
  setSelectedProjectId: (selectedProjectId) => set({ selectedProjectId }),
  activeBoardId: null,
  setActiveBoardId: (activeBoardId) => set({ activeBoardId }),
  isBoardsPanelCollapsed:
    typeof window !== 'undefined'
      ? localStorage.getItem('takt_boards_panel_collapsed') === 'true'
      : false,
  setBoardsPanelCollapsed: (collapsed) => {
    localStorage.setItem('takt_boards_panel_collapsed', String(collapsed));
    set({ isBoardsPanelCollapsed: collapsed });
  },
  toggleBoardsPanel: () =>
    set((state) => {
      const next = !state.isBoardsPanelCollapsed;
      localStorage.setItem('takt_boards_panel_collapsed', String(next));
      return { isBoardsPanelCollapsed: next };
    }),
  viewMode: 'kanban',
  setViewMode: (viewMode) => set({ viewMode }),
  selectedDateFilter: null,
  setSelectedDateFilter: (selectedDateFilter) => set({ selectedDateFilter }),
}));
