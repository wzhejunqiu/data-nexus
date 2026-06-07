import { describe, expect, it } from 'vitest'
import i18n from '@/i18n'
import { isDialogCancelled, mapWailsError, translateAppError } from './errors'

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

  it('translates known error codes', () => {
    const t = i18n.getFixedT('en')
    const msg = translateAppError(t, { code: 'READ_ONLY', message: 'fallback' })
    expect(msg).toContain('read-only')
  })

  it('falls back to message for unknown codes', () => {
    const t = i18n.getFixedT('en')
    const msg = translateAppError(t, { code: 'UNKNOWN', message: 'custom' })
    expect(msg).toBe('custom')
  })
})
