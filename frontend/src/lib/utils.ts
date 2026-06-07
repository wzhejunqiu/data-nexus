import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'
import type { CSVFormatOptions } from './types/csv'
import { defaultCSVFormat } from './types/csv'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatCell(value: unknown): string {
  if (value === null || value === undefined) return 'NULL'
  if (
    typeof value === 'object' &&
    value !== null &&
    'type' in value &&
    (value as { type: string }).type === 'blob'
  ) {
    const blob = value as { type: string; size: number }
    return `[BLOB ${blob.size} bytes]`
  }
  return String(value)
}

function escapeCSVField(s: string, quoteChar: string): string {
  const q = quoteChar || '"'
  const needsQuote = s.includes(q) || s.includes('\n') || s.includes('\r') || s.includes(',') || s.includes('\t') || s.includes(';')
  if (needsQuote) {
    const escaped = s.split(q).join(q + q)
    return `${q}${escaped}${q}`
  }
  return s
}

export function rowsToCSV(
  columns: string[],
  rows: Record<string, unknown>[],
  format?: Partial<CSVFormatOptions>,
): string {
  const opts = { ...defaultCSVFormat(), ...format }
  const delimiter = opts.delimiter || ','
  const quoteChar = opts.quoteChar || '"'
  const lineEnding = opts.lineEnding === 'lf' ? '\n' : '\r\n'

  const lines: string[] = []
  if (opts.hasHeader) {
    lines.push(columns.map((c) => escapeCSVField(c, quoteChar)).join(delimiter))
  }
  for (const row of rows) {
    lines.push(
      columns
        .map((col) => {
          const val = row[col]
          if (val === null || val === undefined) return ''
          return escapeCSVField(formatCell(val), quoteChar)
        })
        .join(delimiter),
    )
  }
  return lines.join(lineEnding)
}

export function serializePK(pk: Record<string, unknown>): string {
  return Object.keys(pk)
    .sort()
    .map((k) => `${k}=${JSON.stringify(pk[k])}`)
    .join('|')
}

export function cellKey(pk: Record<string, unknown>, columnName: string): string {
  return `${serializePK(pk)}:${columnName}`
}

export function isBlobValue(value: unknown): boolean {
  return (
    typeof value === 'object' &&
    value !== null &&
    'type' in value &&
    (value as { type: string }).type === 'blob'
  )
}

export function pkSummary(pk: Record<string, unknown>): string {
  return Object.entries(pk)
    .map(([k, v]) => `${k}=${formatCell(v)}`)
    .join(', ')
}
