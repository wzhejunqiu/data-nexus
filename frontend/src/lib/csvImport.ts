export const SQL_COLUMN_TYPES = ['INTEGER', 'REAL', 'TEXT', 'BLOB'] as const
export type SqlColumnType = (typeof SQL_COLUMN_TYPES)[number]

export interface NewTableColumnSpec {
  dataType: SqlColumnType
  primaryKey: boolean
}

export function inferColumnType(samples: string[]): SqlColumnType {
  for (const raw of samples) {
    const v = raw.trim()
    if (v === '') continue
    if (/^-?\d+$/.test(v)) return 'INTEGER'
    if (/^-?\d+(\.\d+)?([eE][+-]?\d+)?$/.test(v)) return 'REAL'
    return 'TEXT'
  }
  return 'TEXT'
}

export function buildNewTableColumnSpecs(
  headers: string[],
  rows: Record<string, unknown>[],
): Record<string, NewTableColumnSpec> {
  const specs: Record<string, NewTableColumnSpec> = {}
  for (const h of headers) {
    const samples = rows.map((row) => String(row[h] ?? ''))
    specs[h] = { dataType: inferColumnType(samples), primaryKey: false }
  }
  return specs
}

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

function quoteIdent(name: string): string {
  return `"${name.replace(/"/g, '""')}"`
}

function quoteJoin(cols: string[]): string {
  return cols.map(quoteIdent).join(', ')
}

export function buildCreateTableSQL(
  tableName: string,
  headers: string[],
  columnMap: Record<string, string>,
  specs: Record<string, NewTableColumnSpec>,
): string | null {
  const entries: { name: string; typ: SqlColumnType; pk: boolean }[] = []
  for (const h of headers) {
    const colName = columnMap[h]
    if (!colName) continue
    const spec = specs[h]
    entries.push({
      name: colName,
      typ: spec?.dataType ?? 'TEXT',
      pk: spec?.primaryKey ?? false,
    })
  }
  if (entries.length === 0) return null

  const pkCols = entries.filter((e) => e.pk).map((e) => e.name)
  const colDefs: string[] = []
  for (const e of entries) {
    if (pkCols.length === 1 && e.pk) {
      colDefs.push(`${quoteIdent(e.name)} ${e.typ} PRIMARY KEY`)
    } else {
      colDefs.push(`${quoteIdent(e.name)} ${e.typ}`)
    }
  }
  if (pkCols.length > 1) {
    colDefs.push(`PRIMARY KEY (${quoteJoin(pkCols)})`)
  }

  const body = colDefs.map((def) => `  ${def}`).join(',\n')
  return `CREATE TABLE ${quoteIdent(tableName)} (\n${body}\n)`
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
