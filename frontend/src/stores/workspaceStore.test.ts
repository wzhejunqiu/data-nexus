import { describe, expect, it, beforeEach } from 'vitest'
import { useWorkspaceStore } from './workspaceStore'

describe('workspaceStore', () => {
  beforeEach(() => {
    useWorkspaceStore.setState({
      activeConnectionId: null,
      selectedTable: null,
      activeTab: 'data',
    })
  })

  it('selectTable sets connection, table, and data tab', () => {
    useWorkspaceStore.getState().selectTable('conn-1', 'users')
    const state = useWorkspaceStore.getState()
    expect(state.activeConnectionId).toBe('conn-1')
    expect(state.selectedTable).toBe('users')
    expect(state.activeTab).toBe('data')
  })

  it('setters update individual fields', () => {
    useWorkspaceStore.getState().setActiveConnectionId('c1')
    useWorkspaceStore.getState().setSelectedTable('orders')
    useWorkspaceStore.getState().setActiveTab('sql')
    const state = useWorkspaceStore.getState()
    expect(state.activeConnectionId).toBe('c1')
    expect(state.selectedTable).toBe('orders')
    expect(state.activeTab).toBe('sql')
  })
})
