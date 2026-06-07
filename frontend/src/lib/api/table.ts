import { BrowseRows } from '../../../wailsjs/go/wails/TableService'
import type { BrowseRowsRequest, PaginatedTableData } from '../types'
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
}
