import { ListQueryHistory, ListSqlExecutions } from '../../../wailsjs/go/wails/SqlExecutionService'
import type { SqlExecutionList, SqlExecutionRecord } from '../types'
import { mapWailsError } from './errors'

async function wrap<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    throw mapWailsError(err)
  }
}

function mapKind(kind: unknown): SqlExecutionRecord['kind'] {
  if (kind === 1 || kind === 'exec') return 'exec'
  return 'result'
}

function mapRecord(raw: {
  id?: number
  connectionId: string
  sql: string
  kind: unknown
  effectRows: number
  durationMs: number
  executedAt: unknown
}): SqlExecutionRecord {
  return {
    id: raw.id,
    connectionId: raw.connectionId,
    sql: raw.sql,
    kind: mapKind(raw.kind),
    effectRows: raw.effectRows,
    durationMs: raw.durationMs,
    executedAt: String(raw.executedAt ?? ''),
  }
}

export const queryHistoryApi = {
  list: (connectionId: string) =>
    wrap(async () => {
      const items = await ListQueryHistory(connectionId)
      return items ?? []
    }),
  listExecutions: (connectionId: string, limit = 50) =>
    wrap(async (): Promise<SqlExecutionList> => {
      const res = await ListSqlExecutions(connectionId, limit)
      const items = (res.items ?? []) as Array<{
        id?: number
        connectionId: string
        sql: string
        kind: unknown
        effectRows: number
        durationMs: number
        executedAt: unknown
      }>
      return { items: items.map((item) => mapRecord(item)) }
    }),
}
