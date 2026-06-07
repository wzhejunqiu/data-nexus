import { CancelExportTableCSV, ExportTableCSV } from '../../../wailsjs/go/wails/ExportService'
import { model } from '../../../wailsjs/go/models'
import type { ExportTableCSVRequest } from '../types'
import { mapWailsError } from './errors'

export const exportApi = {
  exportTableCSV: async (req: ExportTableCSVRequest): Promise<string> => {
    try {
      return (await ExportTableCSV(model.ExportTableCSVRequest.createFrom(req))) as string
    } catch (err) {
      throw mapWailsError(err)
    }
  },
  cancelExportTableCSV: async (exportId: string): Promise<void> => {
    try {
      await CancelExportTableCSV(exportId)
    } catch (err) {
      throw mapWailsError(err)
    }
  },
}
