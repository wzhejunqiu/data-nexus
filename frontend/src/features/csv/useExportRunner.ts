import { exportApi } from '@/lib/api/export'
import { formatError, isDialogCancelled, isExportCancelled, mapWailsError } from '@/lib/api/errors'
import { fileApi } from '@/lib/api/file'
import type { CSVFormatOptions } from '@/lib/types/csv'
import { saveCSVFormatPreference } from '@/lib/types/csv'
import { rowsToCSV } from '@/lib/utils'
import type { ExportResult, ExportSource } from '@/stores/exportStore'
import type { TFunction } from 'i18next'

export interface RunExportParams {
  source: ExportSource
  connectionId: string
  tableName?: string
  columns: string[]
  format: CSVFormatOptions
  filePath: string
  rows?: Record<string, unknown>[]
  exportId: string | null
  t: TFunction
}

export async function runExport(params: RunExportParams): Promise<ExportResult> {
  const { source, connectionId, tableName, columns, format, filePath, rows, exportId, t } = params
  const start = performance.now()
  saveCSVFormatPreference(format)

  try {
    if (source === 'table-all') {
      await exportApi.exportTableCSV({
        connectionId,
        tableName: tableName!,
        format,
        columns,
        defaultPath: filePath,
        exportId: exportId ?? undefined,
      })
      return {
        status: 'success',
        filePath,
        columnCount: columns.length,
        durationMs: Math.round(performance.now() - start),
      }
    }

    const csv = rowsToCSV(columns, rows ?? [], format)
    await fileApi.writeTextFile(filePath, csv, format.encoding)
    return {
      status: 'success',
      filePath,
      rowsExported: rows?.length ?? 0,
      columnCount: columns.length,
      durationMs: Math.round(performance.now() - start),
    }
  } catch (err) {
    const appErr = mapWailsError(err)
    if (isExportCancelled(appErr)) {
      return {
        status: 'cancelled',
        message: t('csv.exportResultCancelled'),
        durationMs: Math.round(performance.now() - start),
      }
    }
    if (isDialogCancelled(appErr)) {
      return {
        status: 'cancelled',
        message: t('csv.exportResultCancelled'),
        durationMs: Math.round(performance.now() - start),
      }
    }
    return {
      status: 'error',
      message: formatError(t, appErr),
      durationMs: Math.round(performance.now() - start),
    }
  }
}
