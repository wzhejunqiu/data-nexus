import type { QueryClient } from '@tanstack/react-query'
import { connectionApi } from '@/lib/api/connection'
import { connectionGroupApi } from '@/lib/api/connectionGroup'
import type { ConnectRequest } from '@/lib/types'

export async function moveNewConnectionToGroup(
  connectionId: string,
  groupId: string | null,
  qc: QueryClient,
) {
  if (!groupId) return
  await connectionGroupApi.moveConnectionToGroup(connectionId, groupId, -1)
  await qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
}

export async function openSqliteAndMaybePlace(
  req: ConnectRequest,
  groupId: string | null,
  qc: QueryClient,
) {
  const conn = await connectionApi.openFromFile(req)
  await moveNewConnectionToGroup(conn.id, groupId, qc)
  await qc.invalidateQueries({ queryKey: ['connections'] })
  return conn
}
