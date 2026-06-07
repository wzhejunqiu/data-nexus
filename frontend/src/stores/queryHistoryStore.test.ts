import { describe, expect, it, beforeEach } from 'vitest'
import { useQueryHistoryStore } from './queryHistoryStore'

describe('queryHistoryStore', () => {
  beforeEach(() => {
    useQueryHistoryStore.setState({ items: {} })
  })

  it('adds sql per connection', () => {
    useQueryHistoryStore.getState().add('c1', 'SELECT 1')
    expect(useQueryHistoryStore.getState().items.c1).toEqual(['SELECT 1'])
  })

  it('deduplicates and moves recent to front', () => {
    const { add } = useQueryHistoryStore.getState()
    add('c1', 'SELECT 1')
    add('c1', 'SELECT 2')
    add('c1', 'SELECT 1')
    expect(useQueryHistoryStore.getState().items.c1).toEqual(['SELECT 1', 'SELECT 2'])
  })

  it('ignores empty sql', () => {
    useQueryHistoryStore.getState().add('c1', '   ')
    expect(useQueryHistoryStore.getState().items.c1).toBeUndefined()
  })

  it('keeps at most 50 entries', () => {
    const { add } = useQueryHistoryStore.getState()
    for (let i = 0; i < 55; i++) {
      add('c1', `SELECT ${i}`)
    }
    expect(useQueryHistoryStore.getState().items.c1).toHaveLength(50)
    expect(useQueryHistoryStore.getState().items.c1[0]).toBe('SELECT 54')
  })

  it('isolates history by connection id', () => {
    useQueryHistoryStore.getState().add('c1', 'SELECT 1')
    useQueryHistoryStore.getState().add('c2', 'SELECT 2')
    expect(useQueryHistoryStore.getState().items.c1).toEqual(['SELECT 1'])
    expect(useQueryHistoryStore.getState().items.c2).toEqual(['SELECT 2'])
  })
})
