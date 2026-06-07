import { create } from 'zustand'

interface StatusState {
  rowCount: number | null
  durationMs: number | null
  operation: string | null
  setStatus: (operation: string, rowCount: number | null, durationMs: number | null) => void
  clear: () => void
}

export const useStatusStore = create<StatusState>((set) => ({
  rowCount: null,
  durationMs: null,
  operation: null,
  setStatus: (operation, rowCount, durationMs) => set({ operation, rowCount, durationMs }),
  clear: () => set({ rowCount: null, durationMs: null, operation: null }),
}))
