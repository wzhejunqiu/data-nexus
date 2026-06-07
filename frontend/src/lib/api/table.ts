import { BrowseRows, UpdateCellsBatch } from '../../../wailsjs/go/wails/TableService'
import { model } from '../../../wailsjs/go/models'
import type {
  BrowseRowsRequest,
  PaginatedTableData,
  UpdateCellsBatchRequest,
  UpdateCellsBatchResult,
} from '../types'
import { mapWailsError } from './errors'

export const tableApi = {
  browseRows: (req: BrowseRowsRequest) =>
    (async () => {
      try {
        return (await BrowseRows(req)) as PaginatedTableData
      } catch (err) {
        throw mapWailsError(err)
      }
    })(),
  updateCellsBatch: async (req: UpdateCellsBatchRequest): Promise<UpdateCellsBatchResult> => {
    try {
      return (await UpdateCellsBatch(
        model.UpdateCellsBatchRequest.createFrom(req),
      )) as UpdateCellsBatchResult
    } catch (err) {
      throw mapWailsError(err)
    }
  },
}
