import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type ThemeMode = 'light' | 'dark' | 'system'

export const THEME_MODES: ThemeMode[] = ['light', 'dark', 'system']

interface ThemeState {
  mode: ThemeMode
  setMode: (mode: ThemeMode) => void
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      mode: 'system',
      setMode: (mode) => set({ mode }),
    }),
    { name: 'data-nexus-theme' },
  ),
)

export function isThemeMode(value: string): value is ThemeMode {
  return THEME_MODES.includes(value as ThemeMode)
}

export function resolveTheme(mode: ThemeMode): 'light' | 'dark' {
  if (mode === 'system') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return mode
}

export function applyThemeMode(mode: ThemeMode) {
  useThemeStore.getState().setMode(mode)
  document.documentElement.classList.toggle('dark', resolveTheme(mode) === 'dark')
}
