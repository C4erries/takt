import { create } from 'zustand';

export interface TaskNavigationState {
  selectedTaskId: string | null;
  setSelectedTaskId: (id: string | null) => void;
  isQuickCaptureOpen: boolean;
  setQuickCaptureOpen: (isOpen: boolean) => void;
  activeFilter: string;
  setActiveFilter: (filter: string) => void;
  activeNav: 'inbox' | 'today' | 'projects';
  setActiveNav: (nav: 'inbox' | 'today' | 'projects') => void;
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
}));
