import { describe, expect, it } from 'vitest'
import i18n from '@/i18n'
import {
  formatError,
  isDialogCancelled,
  isExportCancelled,
  mapWailsError,
  translateAppError,
} from './errors'

const KNOWN_ERROR_CODES = [
  'INVALID_REQUEST',
  'INVALID_PATH',
  'CONNECTION_NOT_FOUND',
  'CONNECTION_ALREADY_OPEN',
  'CONNECTION_FAILED',
  'DATABASE_LOCKED',
  'TABLE_NOT_FOUND',
  'SQL_ERROR',
  'RESULT_TOO_LARGE',
  'READ_ONLY',
  'DIALOG_CANCELLED',
  'EXPORT_CANCELLED',
  'EXPORT_NO_STABLE_KEY',
  'SAVED_NOT_FOUND',
  'INTERNAL_ERROR',
] as const

describe('mapWailsError', () => {
  it.each(KNOWN_ERROR_CODES)('parses object shape for %s', (code) => {
    const err = mapWailsError({ code, message: `${code} detail` })
    expect(err.code).toBe(code)
    expect(err.message).toBe(`${code} detail`)
  })

  it.each(KNOWN_ERROR_CODES)('parses string shape for %s', (code) => {
    const err = mapWailsError(`${code}: ${code} detail`)
    expect(err.code).toBe(code)
    expect(err.message).toBe(`${code} detail`)
  })

  it('detects dialog cancelled', () => {
    const err = mapWailsError({ code: 'DIALOG_CANCELLED', message: 'cancelled' })
    expect(isDialogCancelled(err)).toBe(true)
  })

  it('detects export cancelled', () => {
    const err = mapWailsError({ code: 'EXPORT_CANCELLED', message: 'export cancelled' })
    expect(isExportCancelled(err)).toBe(true)
  })

  it('falls back to INTERNAL_ERROR for unknown codes', () => {
    const err = mapWailsError('UNKNOWN_CODE: something broke')
    expect(err.code).toBe('INTERNAL_ERROR')
    expect(err.message).toBe('UNKNOWN_CODE: something broke')
  })

  it('shows SQL error message instead of generic internal text', () => {
    const t = i18n.getFixedT('zh-CN')
    const msg = translateAppError(t, { code: 'SQL_ERROR', message: 'near DESC: syntax error' })
    expect(msg).toBe('near DESC: syntax error')
  })

  it.each([
    ['CONNECTION_FAILED', 'en', 'Failed to open database'],
    ['DATABASE_LOCKED', 'en', 'Database is locked'],
    ['RESULT_TOO_LARGE', 'en', 'row limit'],
    ['READ_ONLY', 'en', 'read-only'],
    ['CONNECTION_FAILED', 'zh-CN', '无法打开'],
    ['DATABASE_LOCKED', 'zh-CN', '数据库被锁定'],
  ] as const)('translates %s in %s', (code, lng, expected) => {
    const t = i18n.getFixedT(lng)
    const msg = translateAppError(t, { code, message: 'fallback' })
    expect(msg.toLowerCase()).toContain(expected.toLowerCase())
  })

  it('falls back to message for unknown codes', () => {
    const t = i18n.getFixedT('en')
    const msg = translateAppError(t, { code: 'UNKNOWN', message: 'custom' })
    expect(msg).toBe('custom')
  })

  it('parses object message string shape', () => {
    const err = mapWailsError({ message: 'SQL_ERROR: syntax error' })
    expect(err.code).toBe('SQL_ERROR')
    expect(err.message).toBe('syntax error')
  })

  it('falls back for object with unknown message code', () => {
    const err = mapWailsError({ message: 'UNKNOWN: broken' })
    expect(err.code).toBe('INTERNAL_ERROR')
    expect(err.message).toBe('UNKNOWN: broken')
  })

  it('stringifies non-object errors', () => {
    const err = mapWailsError(404)
    expect(err.code).toBe('INTERNAL_ERROR')
    expect(err.message).toBe('404')
  })

  it('includes details when provided on object shape', () => {
    const err = mapWailsError({
      code: 'INVALID_REQUEST',
      message: 'bad',
      details: { field: 'sql' },
    })
    expect(err.details).toEqual({ field: 'sql' })
  })

  it('formatError maps and translates end to end', () => {
    const t = i18n.getFixedT('en')
    const msg = formatError(t, { code: 'READ_ONLY', message: 'fallback' })
    expect(msg.toLowerCase()).toContain('read-only')
  })

  it('translateAppError uses internal fallback when no message', () => {
    const t = i18n.getFixedT('en')
    const msg = translateAppError(t, { code: 'TOTALLY_UNKNOWN', message: '' })
    expect(msg.length).toBeGreaterThan(0)
  })
})
