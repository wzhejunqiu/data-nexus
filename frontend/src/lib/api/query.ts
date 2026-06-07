import { Execute, ClassifySQL } from '../../../wailsjs/go/wails/QueryService'
import type { ExecuteQueryRequest, QueryResponse, StatementKind } from '../types'
import { mapWailsError } from './errors'

export const queryApi = {
  classifySQL: (connectionId: string, sql: string) =>
    (async () => {
      try {
        return (await ClassifySQL(connectionId, sql)) as StatementKind
      } catch (err) {
        throw mapWailsError(err)
      }
    })(),
  execute: (req: ExecuteQueryRequest) =>
    (async () => {
      try {
        return (await Execute({
          ...req,
          params: req.params ?? [],
          maxRows: req.maxRows ?? 10000,
        })) as QueryResponse
      } catch (err) {
        throw mapWailsError(err)
      }
    })(),
}
