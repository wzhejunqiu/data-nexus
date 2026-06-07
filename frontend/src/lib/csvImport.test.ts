import { describe, expect, it } from 'vitest'
import { buildImportColumnMap, sanitizeColumnName } from './csvImport'

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
