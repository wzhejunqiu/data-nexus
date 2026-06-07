import { Execute } from '../../../wailsjs/go/wails/QueryService'
import type { ExecuteQueryRequest, QueryResponse } from '../types'
import { mapWailsError } from './errors'

export const queryApi = {
  execute: (req: ExecuteQueryRequest) =>
    (async () => {
      try {
        return (await Execute({
          ...req,
          params: req.params ?? [],
          maxRows: req.maxRows ?? 1000,
        })) as QueryResponse
      } catch (err) {
        throw mapWailsError(err)
      }
    })(),
}
