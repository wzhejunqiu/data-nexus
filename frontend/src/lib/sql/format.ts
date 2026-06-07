import { format as formatSQL } from 'sql-formatter'
import type { SqlDialect } from './dialect'

export function formatSql(sql: string, dialect: SqlDialect): string {
  return formatSQL(sql, { language: dialect })
}

export function formatSqlSafe(sql: string, dialect: SqlDialect): string {
  try {
    return formatSql(sql, dialect)
  } catch {
    return sql
  }
}
