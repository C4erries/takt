import { create } from 'zustand';
import { parseUrl, LAST_ROUTE_KEY } from './urlParser';

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

function getInitialRoute() {
  if (typeof window === 'undefined') {
    return { activeNav: 'inbox', selectedProjectId: null, activeBoardId: null, selectedTaskId: null };
  }
  let pathname = window.location.pathname;
  let search = window.location.search;

  if (pathname === '/' || pathname === '') {
    const saved = localStorage.getItem(LAST_ROUTE_KEY);
    if (saved && saved !== '/' && saved !== '') {
      window.history.replaceState(null, '', saved);
      pathname = window.location.pathname;
      search = window.location.search;
    }
  }

  return parseUrl(pathname, search);
}

const initialRoute = getInitialRoute();

export const useTaskStore = create<TaskNavigationState>((set) => ({
  selectedTaskId: initialRoute.selectedTaskId,
  setSelectedTaskId: (selectedTaskId) => set({ selectedTaskId }),
  isQuickCaptureOpen: false,
  setQuickCaptureOpen: (isQuickCaptureOpen) => set({ isQuickCaptureOpen }),
  activeFilter: 'all',
  setActiveFilter: (activeFilter) => set({ activeFilter }),
  activeNav: initialRoute.activeNav,
  setActiveNav: (activeNav) => set({ activeNav }),
  selectedProjectId: initialRoute.selectedProjectId,
  setSelectedProjectId: (selectedProjectId) => set({ selectedProjectId }),
  activeBoardId: initialRoute.activeBoardId,
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
