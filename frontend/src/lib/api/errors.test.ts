import { describe, expect, it } from 'vitest'
import { isDialogCancelled, mapWailsError } from './errors'

describe('mapWailsError', () => {
  it('parses app error shape', () => {
    const err = mapWailsError({ code: 'SQL_ERROR', message: 'syntax error' })
    expect(err.code).toBe('SQL_ERROR')
    expect(err.message).toBe('syntax error')
  })

  it('detects dialog cancelled', () => {
    const err = mapWailsError({ code: 'DIALOG_CANCELLED', message: 'cancelled' })
    expect(isDialogCancelled(err)).toBe(true)
  })
})
