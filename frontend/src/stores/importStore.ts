import { create } from 'zustand'

export interface ImportSession {
  connectionId: string
  defaultTable?: string
}

export interface ImportResult {
  status: 'success' | 'error'
  rowsInserted?: number
  rowsUpdated?: number
  durationMs?: number
  message?: string
}

interface ImportState {
  session: ImportSession | null
  openImport: (session: ImportSession) => void
  closeImport: () => void
}

export const useImportStore = create<ImportState>((set) => ({
  session: null,
  openImport: (session) => set({ session }),
  closeImport: () => set({ session: null }),
}))
