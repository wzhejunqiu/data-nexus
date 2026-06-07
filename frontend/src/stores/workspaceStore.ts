import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { RowFilter } from '@/lib/types'

export interface TableBrowseState {
  filters: RowFilter[]
  sort: string
  order: 'asc' | 'desc'
  search: string
}

const defaultTableState = (): TableBrowseState => ({
  filters: [],
  sort: '',
  order: 'asc',
  search: '',
})

interface WorkspaceState {
  activeConnectionId: string | null
  selectedTable: string | null
  activeTab: 'schema' | 'data' | 'sql'
  pendingSql: string | null
  editorSql: string
  tableStates: Record<string, Record<string, TableBrowseState>>
  setActiveConnectionId: (id: string | null) => void
  setSelectedTable: (table: string | null) => void
  setActiveTab: (tab: 'schema' | 'data' | 'sql') => void
  setPendingSql: (sql: string | null) => void
  setEditorSql: (sql: string) => void
  consumePendingSql: () => string | null
  selectTable: (connectionId: string, table: string) => void
  getTableState: (connectionId: string, table: string) => TableBrowseState
  setTableState: (connectionId: string, table: string, patch: Partial<TableBrowseState>) => void
}

export const useWorkspaceStore = create<WorkspaceState>()(
  persist(
    (set, get) => ({
      activeConnectionId: null,
      selectedTable: null,
      activeTab: 'data',
      pendingSql: null,
      editorSql: 'SELECT 1;',
      tableStates: {},
      setActiveConnectionId: (id) => set({ activeConnectionId: id }),
      setSelectedTable: (table) => set({ selectedTable: table }),
      setActiveTab: (tab) => set({ activeTab: tab }),
      setPendingSql: (sql) => set({ pendingSql: sql }),
      setEditorSql: (sql) => set({ editorSql: sql }),
      consumePendingSql: () => {
        const sql = get().pendingSql
        if (sql) set({ pendingSql: null })
        return sql
      },
      selectTable: (connectionId, table) =>
        set({ activeConnectionId: connectionId, selectedTable: table, activeTab: 'data' }),
      getTableState: (connectionId, table) => {
        const conn = get().tableStates[connectionId]
        return conn?.[table] ?? defaultTableState()
      },
      setTableState: (connectionId, table, patch) => {
        const all = { ...get().tableStates }
        const conn = { ...(all[connectionId] ?? {}) }
        conn[table] = { ...(conn[table] ?? defaultTableState()), ...patch }
        all[connectionId] = conn
        set({ tableStates: all })
      },
    }),
    {
      name: 'data-nexus:workspace',
      partialize: (state) => ({
        activeConnectionId: state.activeConnectionId,
        selectedTable: state.selectedTable,
        activeTab: state.activeTab,
        tableStates: state.tableStates,
      }),
    },
  ),
)
