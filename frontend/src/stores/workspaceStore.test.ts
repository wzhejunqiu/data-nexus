import { describe, expect, it, beforeEach } from 'vitest'
import { useWorkspaceStore } from './workspaceStore'

describe('workspaceStore', () => {
  beforeEach(() => {
    useWorkspaceStore.setState({
      activeConnectionId: null,
      selectedTable: null,
      activeTab: 'data',
      pendingSql: null,
      editorSql: 'SELECT 1;',
      tableStates: {},
    })
  })

  it('selectTable sets connection, table, and data tab', () => {
    useWorkspaceStore.getState().selectTable('conn-1', 'users')
    const state = useWorkspaceStore.getState()
    expect(state.activeConnectionId).toBe('conn-1')
    expect(state.selectedTable).toBe('users')
    expect(state.activeTab).toBe('data')
  })

  it('persists table browse state per connection', () => {
    useWorkspaceStore.getState().setTableState('c1', 'users', {
      filters: [{ column: 'id', operator: 'eq', value: '1' }],
      sort: 'id',
      order: 'desc',
    })
    const s = useWorkspaceStore.getState().getTableState('c1', 'users')
    expect(s.filters).toHaveLength(1)
    expect(s.sort).toBe('id')
    expect(s.order).toBe('desc')
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

  it('setEditorSql updates editor content', () => {
    useWorkspaceStore.getState().setEditorSql('SELECT 99;')
    expect(useWorkspaceStore.getState().editorSql).toBe('SELECT 99;')
  })
})
