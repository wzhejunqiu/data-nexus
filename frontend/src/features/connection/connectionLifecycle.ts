import type { QueryClient } from '@tanstack/react-query'
import { connectionApi } from '@/lib/api/connection'
import { useWorkspaceStore } from '@/stores/workspaceStore'

export function invalidateConnectionQueries(qc: QueryClient) {
  void qc.invalidateQueries({ queryKey: ['connections'] })
  void qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
}

export async function closeActiveConnection(qc: QueryClient, connectionId: string | null) {
  if (!connectionId) return
  await connectionApi.close(connectionId)
  useWorkspaceStore.getState().setActiveConnectionId(null)
  useWorkspaceStore.getState().setSelectedTable(null)
  invalidateConnectionQueries(qc)
}
