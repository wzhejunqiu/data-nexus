import { describe, expect, it } from 'vitest'
import { formatCell } from '@/lib/utils'

describe('DataGrid cell rendering helpers', () => {
  it('formats null as NULL string in helper', () => {
    expect(formatCell(null)).toBe('NULL')
  })

  it('formats blob cells', () => {
    expect(formatCell({ type: 'blob', size: 3 })).toBe('[BLOB 3 bytes]')
  })
})
