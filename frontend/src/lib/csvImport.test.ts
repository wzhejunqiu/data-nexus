import { describe, expect, it } from 'vitest'
import {
  buildCreateTableSQL,
  buildImportColumnMap,
  buildNewTableColumnSpecs,
  inferColumnType,
  sanitizeColumnName,
} from './csvImport'

describe('sanitizeColumnName', () => {
  it('replaces invalid characters with underscores', () => {
    expect(sanitizeColumnName('user-id')).toBe('user_id')
  })

  it('prefixes numeric-leading names', () => {
    expect(sanitizeColumnName('123')).toBe('col_123')
  })
})

describe('buildImportColumnMap', () => {
  it('sanitizes headers for new tables', () => {
    expect(buildImportColumnMap(['User ID', 'name'])).toEqual({
      'User ID': 'User_ID',
      name: 'name',
    })
  })

  it('matches target columns by exact name', () => {
    expect(buildImportColumnMap(['ID', 'Name'], ['id', 'name'])).toEqual({
      ID: 'id',
      Name: 'name',
    })
  })

  it('matches target columns after sanitizing header', () => {
    expect(buildImportColumnMap(['user-id'], ['user_id'])).toEqual({
      'user-id': 'user_id',
    })
  })

  it('leaves unmatched headers empty for existing tables', () => {
    expect(buildImportColumnMap(['unknown'], ['id'])).toEqual({
      unknown: '',
    })
  })
})

describe('inferColumnType', () => {
  it('detects integers and reals', () => {
    expect(inferColumnType(['42', '3.14', 'hello'])).toBe('INTEGER')
    expect(inferColumnType(['', '2.5'])).toBe('REAL')
    expect(inferColumnType(['', ''])).toBe('TEXT')
  })
})

describe('buildCreateTableSQL', () => {
  it('builds single-column primary key inline', () => {
    expect(
      buildCreateTableSQL(
        'users',
        ['id', 'name'],
        { id: 'item_id', name: 'title' },
        {
          id: { dataType: 'INTEGER', primaryKey: true },
          name: { dataType: 'TEXT', primaryKey: false },
        },
      ),
    ).toBe(
      'CREATE TABLE "users" (\n  "item_id" INTEGER PRIMARY KEY,\n  "title" TEXT\n)',
    )
  })

  it('builds composite primary key constraint', () => {
    expect(
      buildCreateTableSQL(
        'pairs',
        ['a', 'b'],
        { a: 'col_a', b: 'col_b' },
        {
          a: { dataType: 'INTEGER', primaryKey: true },
          b: { dataType: 'INTEGER', primaryKey: true },
        },
      ),
    ).toBe(
      'CREATE TABLE "pairs" (\n  "col_a" INTEGER,\n  "col_b" INTEGER,\n  PRIMARY KEY ("col_a", "col_b")\n)',
    )
  })

  it('returns null when no columns are mapped', () => {
    expect(buildCreateTableSQL('t', ['a'], { a: '' }, {})).toBeNull()
  })
})

describe('buildNewTableColumnSpecs', () => {
  it('infers types from preview rows', () => {
    expect(
      buildNewTableColumnSpecs(['id', 'score', 'name'], [
        { id: '1', score: '9.5', name: 'alice' },
      ]),
    ).toEqual({
      id: { dataType: 'INTEGER', primaryKey: false },
      score: { dataType: 'REAL', primaryKey: false },
      name: { dataType: 'TEXT', primaryKey: false },
    })
  })
})
