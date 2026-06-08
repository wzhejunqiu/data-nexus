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

interface BrowseContext {
  database?: string
  schema?: string
}

interface WorkspaceState {
  activeConnectionId: string | null
  selectedTable: string | null
  activeTab: 'schema' | 'data' | 'sql'
  pendingSql: string | null
  editorSql: string
  browseContext: Record<string, BrowseContext>
  tableStates: Record<string, Record<string, TableBrowseState>>
  setActiveConnectionId: (id: string | null) => void
  setSelectedTable: (table: string | null) => void
  setActiveTab: (tab: 'schema' | 'data' | 'sql') => void
  setPendingSql: (sql: string | null) => void
  setEditorSql: (sql: string) => void
  setBrowseContext: (connectionId: string, ctx: BrowseContext) => void
  getBrowseContext: (connectionId: string) => BrowseContext
  consumePendingSql: () => string | null
  selectTable: (connectionId: string, table: string, ctx?: BrowseContext) => void
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
      browseContext: {},
      tableStates: {},
      setActiveConnectionId: (id) => set({ activeConnectionId: id }),
      setSelectedTable: (table) => set({ selectedTable: table }),
      setActiveTab: (tab) => set({ activeTab: tab }),
      setPendingSql: (sql) => set({ pendingSql: sql }),
      setEditorSql: (sql) => set({ editorSql: sql }),
      setBrowseContext: (connectionId, ctx) =>
        set((s) => ({
          browseContext: { ...s.browseContext, [connectionId]: ctx },
        })),
      getBrowseContext: (connectionId) => get().browseContext[connectionId] ?? {},
      consumePendingSql: () => {
        const sql = get().pendingSql
        if (sql) set({ pendingSql: null })
        return sql
      },
      selectTable: (connectionId, table, ctx) => {
        const patch: Partial<WorkspaceState> = {
          activeConnectionId: connectionId,
          selectedTable: table,
          activeTab: 'data',
        }
        if (ctx) {
          patch.browseContext = { ...get().browseContext, [connectionId]: ctx }
        }
        set(patch)
      },
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
