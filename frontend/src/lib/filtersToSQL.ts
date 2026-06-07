import type { RowFilter } from './types'

function escapeSQLString(value: string): string {
  return value.replace(/'/g, "''")
}

function quoteIdent(name: string): string {
  return `"${name.replace(/"/g, '""')}"`
}

function filterToExpr(filter: RowFilter): string {
  const col = quoteIdent(filter.column)
  switch (filter.operator) {
    case 'eq':
      return `${col} = '${escapeSQLString(filter.value ?? '')}'`
    case 'ne':
      return `${col} != '${escapeSQLString(filter.value ?? '')}'`
    case 'gt':
      return `${col} > '${escapeSQLString(filter.value ?? '')}'`
    case 'gte':
      return `${col} >= '${escapeSQLString(filter.value ?? '')}'`
    case 'lt':
      return `${col} < '${escapeSQLString(filter.value ?? '')}'`
    case 'lte':
      return `${col} <= '${escapeSQLString(filter.value ?? '')}'`
    case 'like': {
      let val = filter.value ?? ''
      if (!val.includes('%') && !val.includes('_')) {
        val = `%${val}%`
      }
      return `${col} LIKE '${escapeSQLString(val)}'`
    }
    case 'is_null':
      return `${col} IS NULL`
    case 'is_not_null':
      return `${col} IS NOT NULL`
    case 'in': {
      const vals = (filter.values ?? []).map((v) => `'${escapeSQLString(v)}'`).join(', ')
      return `${col} IN (${vals})`
    }
    default:
      return ''
  }
}

export function filtersToSQL(
  tableName: string,
  filters: RowFilter[],
  sort?: string,
  order?: 'asc' | 'desc',
): string {
  const tableRef = tableName.includes('.')
    ? tableName
        .split('.')
        .map((p) => quoteIdent(p))
        .join('.')
    : quoteIdent(tableName)
  const parts = filters.map(filterToExpr).filter(Boolean)
  let sql = `SELECT * FROM ${tableRef}`
  if (parts.length > 0) {
    sql += ` WHERE ${parts.join(' AND ')}`
  }
  if (sort) {
    sql += ` ORDER BY ${quoteIdent(sort)} ${order === 'desc' ? 'DESC' : 'ASC'}`
  }
  sql += ';'
  return sql
}
