export interface CSVFormatOptions {
  delimiter: string
  quoteChar: string
  hasHeader: boolean
  lineEnding: string
  encoding: string
  commentChar?: string
  lazyQuotes?: boolean
  trimLeadingSpace?: boolean
}

export interface FileFilter {
  displayName: string
  pattern: string
}

export const CSV_EXPORT_WARN_ROWS = 10_000

export interface ExportTableCSVRequest {
  connectionId: string
  tableName: string
  format: CSVFormatOptions
  columns?: string[]
  defaultPath?: string
  exportId?: string
}

export interface ParseCSVPreviewRequest {
  filePath: string
  maxRows: number
  format: CSVFormatOptions
}

export interface CSVPreview {
  headers: string[]
  rows: Record<string, unknown>[]
  rowCount: number
  hasHeader: boolean
}

export interface ImportCSVRequest {
  connectionId: string
  targetTable: string
  newTableName: string
  mode: 'append' | 'update'
  columnMap: Record<string, string>
  filePath: string
  upsertKeys: string[]
  format: CSVFormatOptions
}

export interface ImportCSVResult {
  rowsInserted: number
  rowsUpdated: number
}

export function defaultCSVFormat(): CSVFormatOptions {
  return {
    delimiter: ',',
    quoteChar: '"',
    hasHeader: true,
    lineEnding: 'crlf',
    encoding: 'utf-8',
    commentChar: '',
    lazyQuotes: false,
    trimLeadingSpace: false,
  }
}

export function loadCSVFormatPreference(): CSVFormatOptions {
  try {
    const raw = localStorage.getItem('data-nexus:csv-format')
    if (raw) return { ...defaultCSVFormat(), ...JSON.parse(raw) }
  } catch {
    /* ignore */
  }
  return defaultCSVFormat()
}

export function saveCSVFormatPreference(format: CSVFormatOptions) {
  localStorage.setItem('data-nexus:csv-format', JSON.stringify(format))
}
