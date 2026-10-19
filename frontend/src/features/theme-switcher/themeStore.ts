import { create } from 'zustand';
import type { ThemeId } from './themes';

export type AmbientIntensity = 'off' | 'subtle' | 'vivid';

export const THEME_STORAGE_KEY = 'takt_theme';
export const AMBIENT_INTENSITY_STORAGE_KEY = 'takt_ambient_intensity';
export const DEFAULT_THEME: ThemeId = 'kyoto';
export const DEFAULT_AMBIENT_INTENSITY: AmbientIntensity = 'subtle';

export interface ThemeStoreState {
  theme: ThemeId;
  ambientIntensity: AmbientIntensity;
  isPickerOpen: boolean;
  setTheme: (theme: ThemeId) => void;
  setAmbientIntensity: (intensity: AmbientIntensity) => void;
  setPickerOpen: (open: boolean) => void;
  togglePicker: () => void;
}

function getInitialTheme(): ThemeId {
  if (typeof window === 'undefined') return DEFAULT_THEME;
  try {
    const stored = localStorage.getItem(THEME_STORAGE_KEY) as ThemeId | null;
    if (stored) return stored;
  } catch {
    // Graceful fallback if localStorage is unavailable
  }
  return DEFAULT_THEME;
}

function getInitialIntensity(): AmbientIntensity {
  if (typeof window === 'undefined') return DEFAULT_AMBIENT_INTENSITY;
  try {
    const stored = localStorage.getItem(AMBIENT_INTENSITY_STORAGE_KEY) as AmbientIntensity | null;
    if (stored && (stored === 'off' || stored === 'subtle' || stored === 'vivid')) {
      return stored;
    }
  } catch {
    // Graceful fallback
  }
  return DEFAULT_AMBIENT_INTENSITY;
}

export const useThemeStore = create<ThemeStoreState>((set) => ({
  theme: getInitialTheme(),
  ambientIntensity: getInitialIntensity(),
  isPickerOpen: false,
  setTheme: (theme: ThemeId) => {
    try {
      localStorage.setItem(THEME_STORAGE_KEY, theme);
    } catch {
      // Graceful fallback
    }
    if (typeof document !== 'undefined') {
      document.documentElement.setAttribute('data-theme', theme);
    }
    set({ theme });
  },
  setAmbientIntensity: (ambientIntensity: AmbientIntensity) => {
    try {
      localStorage.setItem(AMBIENT_INTENSITY_STORAGE_KEY, ambientIntensity);
    } catch {
      // Graceful fallback
    }
    set({ ambientIntensity });
  },
  setPickerOpen: (isPickerOpen: boolean) => set({ isPickerOpen }),
  togglePicker: () => set((state) => ({ isPickerOpen: !state.isPickerOpen })),
}));

