import { ImportCSV, ParseCSVPreview } from '../../../wailsjs/go/wails/ImportService'
import { model } from '../../../wailsjs/go/models'
import type {
  CSVPreview,
  ImportCSVRequest,
  ImportCSVResult,
  ParseCSVPreviewRequest,
} from '../types'
import { mapWailsError } from './errors'

export const importApi = {
  parseCSVPreview: async (req: ParseCSVPreviewRequest): Promise<CSVPreview> => {
    try {
      return (await ParseCSVPreview(model.ParseCSVPreviewRequest.createFrom(req))) as CSVPreview
    } catch (err) {
      throw mapWailsError(err)
    }
  },
  importCSV: async (req: ImportCSVRequest): Promise<ImportCSVResult> => {
    try {
      return (await ImportCSV(model.ImportCSVRequest.createFrom(req))) as ImportCSVResult
    } catch (err) {
      throw mapWailsError(err)
    }
  },
}
