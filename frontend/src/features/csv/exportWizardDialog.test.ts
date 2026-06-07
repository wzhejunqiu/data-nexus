import { describe, expect, it } from 'vitest'
import { isDialogCancelled, mapWailsError } from '@/lib/api/errors'

describe('dialog cancel handling', () => {
  it('recognizes DIALOG_CANCELLED from saveFile errors', () => {
    const err = mapWailsError({ code: 'DIALOG_CANCELLED', message: 'cancelled' })
    expect(isDialogCancelled(err)).toBe(true)
  })

  it('does not treat other errors as cancelled', () => {
    const err = mapWailsError({ code: 'INTERNAL_ERROR', message: 'boom' })
    expect(isDialogCancelled(err)).toBe(false)
  })
})
