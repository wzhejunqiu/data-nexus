import { describe, expect, it, beforeEach } from 'vitest'
import { useStatusStore } from './statusStore'

describe('statusStore', () => {
  beforeEach(() => {
    useStatusStore.setState({ rowCount: null, durationMs: null, operation: null })
  })

  it('setStatus writes operation metrics', () => {
    useStatusStore.getState().setStatus('query', 10, 42)
    const state = useStatusStore.getState()
    expect(state.operation).toBe('query')
    expect(state.rowCount).toBe(10)
    expect(state.durationMs).toBe(42)
  })

  it('clear resets all fields', () => {
    useStatusStore.getState().setStatus('exec', 1, 5)
    useStatusStore.getState().clear()
    const state = useStatusStore.getState()
    expect(state.operation).toBeNull()
    expect(state.rowCount).toBeNull()
    expect(state.durationMs).toBeNull()
  })
})
