import type { QueryClient } from '@tanstack/react-query'
import { connectionApi } from '@/lib/api/connection'
import { useWorkspaceStore } from '@/stores/workspaceStore'

export function invalidateConnectionQueries(qc: QueryClient) {
  void qc.invalidateQueries({ queryKey: ['connections'] })
  void qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
}

export function invalidateAttachQueries(qc: QueryClient, connectionId: string) {
  void qc.invalidateQueries({ queryKey: ['namespaces', connectionId] })
  void qc.invalidateQueries({ queryKey: ['tables', connectionId] })
  void qc.invalidateQueries({ queryKey: ['attached', connectionId] })
  invalidateConnectionQueries(qc)
}

export async function closeActiveConnection(qc: QueryClient, connectionId: string | null) {
  if (!connectionId) return
  await connectionApi.close(connectionId)
  useWorkspaceStore.getState().setActiveConnectionId(null)
  useWorkspaceStore.getState().setSelectedTable(null)
  invalidateConnectionQueries(qc)
}
