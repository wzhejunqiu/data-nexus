import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

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

export function isWriteSQL(sql: string): boolean {
  const trimmed = sql.trim().toUpperCase()
  return !(
    trimmed.startsWith('SELECT') ||
    trimmed.startsWith('WITH') ||
    trimmed.startsWith('PRAGMA') ||
    trimmed.startsWith('EXPLAIN') ||
    trimmed.startsWith('DESC') ||
    trimmed.startsWith('DESCRIBE')
  )
}

export function rowsToCSV(columns: string[], rows: Record<string, unknown>[]): string {
  const header = columns.join(',')
  const body = rows
    .map((row) =>
      columns
        .map((col) => {
          const val = row[col]
          if (val === null || val === undefined) return ''
          const s = formatCell(val)
          return s.includes(',') || s.includes('"') ? `"${s.replace(/"/g, '""')}"` : s
        })
        .join(','),
    )
    .join('\n')
  return `${header}\n${body}`
}
