import {
  DeleteCannedQuery,
  ListCannedQueries,
  SaveCannedQuery,
} from '../../../wailsjs/go/wails/CannedQueryService'
import { model } from '../../../wailsjs/go/models'
import type { CannedQuery, CannedQueryList, SaveCannedQueryRequest } from '../types'
import { mapWailsError } from './errors'

async function wrap<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    throw mapWailsError(err)
  }
}

export const queriesApi = {
  list: () => wrap(() => ListCannedQueries() as Promise<CannedQueryList>),
  save: (req: SaveCannedQueryRequest) =>
    wrap(
      () => SaveCannedQuery(model.SaveCannedQueryRequest.createFrom(req)) as Promise<CannedQuery>,
    ),
  delete: (id: string) => wrap(() => DeleteCannedQuery(id)),
}
