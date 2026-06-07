import { create } from 'zustand'

interface QueryHistoryState {
  items: Record<string, string[]>
  add: (connectionId: string, sql: string) => void
}

const MAX_HISTORY = 50

export const useQueryHistoryStore = create<QueryHistoryState>((set, get) => ({
  items: {},
  add: (connectionId, sql) => {
    const trimmed = sql.trim()
    if (!trimmed) return
    const current = get().items[connectionId] ?? []
    const next = [trimmed, ...current.filter((s) => s !== trimmed)].slice(0, MAX_HISTORY)
    set({ items: { ...get().items, [connectionId]: next } })
  },
}))
