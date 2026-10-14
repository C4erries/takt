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
  viewMode: 'kanban',
  setViewMode: (viewMode) => set({ viewMode }),
  selectedDateFilter: null,
  setSelectedDateFilter: (selectedDateFilter) => set({ selectedDateFilter }),
}));
