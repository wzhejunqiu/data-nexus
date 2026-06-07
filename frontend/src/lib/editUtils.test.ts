import { describe, expect, it } from 'vitest'
import { valuesEqual, coerceNewValue } from './utils'

describe('valuesEqual', () => {
  it('treats null and empty as equal', () => {
    expect(valuesEqual(null, '')).toBe(true)
    expect(valuesEqual(undefined, null)).toBe(true)
  })

  it('compares number and numeric string', () => {
    expect(valuesEqual(1, '1')).toBe(true)
    expect(valuesEqual(1, '2')).toBe(false)
  })
})

describe('coerceNewValue', () => {
  it('returns null for empty edit', () => {
    expect(coerceNewValue('x', '', 'TEXT')).toBe(null)
  })

  it('parses integers for INTEGER columns', () => {
    expect(coerceNewValue(1, '42', 'INTEGER')).toBe(42)
  })
})
