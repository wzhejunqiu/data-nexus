import { describe, expect, it } from 'vitest'
import { cn, formatCell, rowsToCSV } from './utils'

describe('cn', () => {
  it('merges class names', () => {
    expect(cn('px-2', 'py-1', false && 'hidden', 'px-4')).toBe('py-1 px-4')
  })
})

describe('formatCell', () => {
  it('renders NULL', () => {
    expect(formatCell(null)).toBe('NULL')
  })

  it('renders undefined as NULL', () => {
    expect(formatCell(undefined)).toBe('NULL')
  })

  it('renders blob placeholder', () => {
    expect(formatCell({ type: 'blob', size: 42 })).toBe('[BLOB 42 bytes]')
  })

  it('renders strings and numbers', () => {
    expect(formatCell('hello')).toBe('hello')
    expect(formatCell(42)).toBe('42')
  })
})

describe('rowsToCSV', () => {
  it('formats header and rows', () => {
    const csv = rowsToCSV(
      ['id', 'name'],
      [
        { id: 1, name: 'alice' },
        { id: 2, name: 'bob' },
      ],
    )
    expect(csv).toBe('id,name\n1,alice\n2,bob')
  })

  it('escapes null as empty field', () => {
    const csv = rowsToCSV(['name'], [{ name: null }])
    expect(csv).toBe('name\n')
  })

  it('quotes fields containing commas', () => {
    const csv = rowsToCSV(['note'], [{ note: 'hello, world' }])
    expect(csv).toBe('note\n"hello, world"')
  })

  it('escapes double quotes', () => {
    const csv = rowsToCSV(['note'], [{ note: 'say "hi"' }])
    expect(csv).toBe('note\n"say ""hi"""')
  })
})
