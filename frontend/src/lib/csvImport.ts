export function sanitizeColumnName(header: string): string {
  let s = header
    .split('')
    .map((ch) => (/^[A-Za-z0-9_]$/.test(ch) ? ch : '_'))
    .join('')
  if (s === '' || /^[0-9]/.test(s)) {
    s = `col_${s}`
  }
  return s
}

export function buildImportColumnMap(
  headers: string[],
  targetColumnNames?: string[],
): Record<string, string> {
  const map: Record<string, string> = {}
  if (!targetColumnNames || targetColumnNames.length === 0) {
    for (const h of headers) {
      map[h] = sanitizeColumnName(h)
    }
    return map
  }

  const byLower = new Map<string, string>()
  const bySanitized = new Map<string, string>()
  for (const col of targetColumnNames) {
    byLower.set(col.toLowerCase(), col)
    bySanitized.set(sanitizeColumnName(col).toLowerCase(), col)
  }

  for (const h of headers) {
    const exact = byLower.get(h.toLowerCase())
    if (exact) {
      map[h] = exact
      continue
    }
    const sanitized = bySanitized.get(sanitizeColumnName(h).toLowerCase())
    map[h] = sanitized ?? ''
  }
  return map
}
