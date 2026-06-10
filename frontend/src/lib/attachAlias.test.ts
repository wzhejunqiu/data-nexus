import { describe, expect, it } from 'vitest'
import { isSafeQuotedIdentifier, validateAttachAlias } from './attachAlias'

describe('attachAlias', () => {
  it('rejects reserved aliases', () => {
    expect(validateAttachAlias('main', [])).toBe('reserved')
    expect(validateAttachAlias('TEMP', [])).toBe('reserved')
  })

  it('rejects invalid characters', () => {
    expect(isSafeQuotedIdentifier('bad"name')).toBe(false)
    expect(validateAttachAlias('bad"name', [])).toBe('invalid')
  })

  it('rejects duplicate aliases', () => {
    expect(validateAttachAlias('logs', ['main', 'logs'])).toBe('duplicate')
  })

  it('accepts valid unique alias', () => {
    expect(validateAttachAlias('analytics', ['main'])).toBeNull()
  })
})
