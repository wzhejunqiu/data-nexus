import { create } from 'zustand'

export type ExportSource = 'table-page' | 'table-all' | 'query-result'

export interface ExportSession {
  source: ExportSource
  connectionId: string
  tableName?: string
  availableColumns: string[]
  rows?: Record<string, unknown>[]
  totalRows?: number
}

export interface ExportResult {
  status: 'success' | 'cancelled' | 'error'
  filePath?: string
  rowsExported?: number
  columnCount?: number
  durationMs?: number
  message?: string
}

interface ExportState {
  session: ExportSession | null
  openExport: (session: ExportSession) => void
  closeExport: () => void
}

export const useExportStore = create<ExportState>((set) => ({
  session: null,
  openExport: (session) => set({ session }),
  closeExport: () => set({ session: null }),
}))
