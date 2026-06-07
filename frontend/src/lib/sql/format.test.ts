import { describe, expect, it } from 'vitest'
import { connectionTypeToSqlDialect } from './dialect'
import { formatSql, formatSqlSafe } from './format'

describe('connectionTypeToSqlDialect', () => {
  it('maps connection types to formatter dialects', () => {
    expect(connectionTypeToSqlDialect('sqlite')).toBe('sqlite')
    expect(connectionTypeToSqlDialect('mysql')).toBe('mysql')
    expect(connectionTypeToSqlDialect('postgresql')).toBe('postgresql')
    expect(connectionTypeToSqlDialect('postgres')).toBe('postgresql')
    expect(connectionTypeToSqlDialect(undefined)).toBe('sql')
  })
})

describe('formatSql', () => {
  const ddl = 'CREATE TABLE "t" ("id" INTEGER PRIMARY KEY, "name" TEXT)'

  it('formats with sqlite dialect', () => {
    const out = formatSql(ddl, 'sqlite')
    expect(out).toContain('CREATE TABLE')
    expect(out).toContain('"id"')
  })

  it('formats with mysql dialect', () => {
    const out = formatSql('CREATE TABLE t (id INT PRIMARY KEY)', 'mysql')
    expect(out).toContain('CREATE TABLE')
  })

  it('formats with postgresql dialect', () => {
    const out = formatSql('CREATE TABLE t (id INTEGER PRIMARY KEY)', 'postgresql')
    expect(out).toContain('CREATE TABLE')
  })
})

describe('formatSqlSafe', () => {
  it('returns original sql when formatting fails', () => {
    const broken = 'CREATE TABLE (((('
    expect(formatSqlSafe(broken, 'sqlite')).toBe(broken)
  })
})
