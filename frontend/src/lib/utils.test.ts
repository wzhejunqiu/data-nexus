import { describe, expect, it } from 'vitest'
import { formatCell } from './utils'

describe('formatCell', () => {
  it('renders NULL', () => {
    expect(formatCell(null)).toBe('NULL')
  })

  it('renders blob placeholder', () => {
    expect(formatCell({ type: 'blob', size: 42 })).toBe('[BLOB 42 bytes]')
  })
})
