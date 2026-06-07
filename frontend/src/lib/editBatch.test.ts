import { describe, expect, it } from 'vitest'
import { isBatchOverLimit, MAX_BATCH_EDITS } from './editBatch'

describe('isBatchOverLimit', () => {
  it('allows up to MAX_BATCH edits', () => {
    expect(isBatchOverLimit(200)).toBe(false)
    expect(isBatchOverLimit(MAX_BATCH_EDITS)).toBe(false)
  })

  it('blocks more than MAX_BATCH edits', () => {
    expect(isBatchOverLimit(201)).toBe(true)
    expect(isBatchOverLimit(MAX_BATCH_EDITS + 1)).toBe(true)
  })
})
