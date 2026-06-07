import { create } from 'zustand'

interface WorkspaceState {
  activeConnectionId: string | null
  selectedTable: string | null
  activeTab: 'schema' | 'data' | 'sql'
  setActiveConnectionId: (id: string | null) => void
  setSelectedTable: (table: string | null) => void
  setActiveTab: (tab: 'schema' | 'data' | 'sql') => void
  selectTable: (connectionId: string, table: string) => void
}

export const useWorkspaceStore = create<WorkspaceState>((set) => ({
  activeConnectionId: null,
  selectedTable: null,
  activeTab: 'data',
  setActiveConnectionId: (id) => set({ activeConnectionId: id }),
  setSelectedTable: (table) => set({ selectedTable: table }),
  setActiveTab: (tab) => set({ activeTab: tab }),
  selectTable: (connectionId, table) =>
    set({ activeConnectionId: connectionId, selectedTable: table, activeTab: 'data' }),
}))
