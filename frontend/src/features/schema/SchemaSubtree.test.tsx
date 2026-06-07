import { describe, expect, it, vi, beforeEach } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { SchemaSubtree } from './SchemaSubtree'
import { useWorkspaceStore } from '@/stores/workspaceStore'

vi.mock('@/lib/api/schema', () => ({
  schemaApi: {
    listTables: vi.fn(),
  },
}))

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    listAttached: vi.fn().mockResolvedValue([]),
    attach: vi.fn(),
    detach: vi.fn(),
  },
}))

vi.mock('@/lib/api/dialog', () => ({
  dialogApi: {
    openDatabaseFile: vi.fn(),
  },
}))

describe('SchemaSubtree', () => {
  beforeEach(() => {
    useWorkspaceStore.setState({
      activeConnectionId: null,
      selectedTable: null,
      activeTab: 'data',
      pendingSql: null,
      tableStates: {},
    })
  })

  it('filters tables and selects one', async () => {
    const { schemaApi } = await import('@/lib/api/schema')
    const onSelectTable = vi.fn()
    vi.mocked(schemaApi.listTables).mockResolvedValue({
      items: [
        { name: 'users', type: 'table', rowCount: 3 },
        { name: 'orders', type: 'table', rowCount: 1 },
        { name: 'v_users', type: 'view' },
      ],
    })

    renderWithProviders(<SchemaSubtree connectionId="c1" onSelectTable={onSelectTable} />)

    await waitFor(() => {
      expect(screen.getByText('users')).toBeInTheDocument()
    })

    fireEvent.change(screen.getByPlaceholderText(/搜索表/), { target: { value: 'order' } })
    expect(screen.queryByText('users')).not.toBeInTheDocument()
    expect(screen.getByText('orders')).toBeInTheDocument()

    fireEvent.click(screen.getByText('orders'))
    expect(onSelectTable).toHaveBeenCalledWith('orders')
  })

  it('applies selected styles without hover override', async () => {
    const { schemaApi } = await import('@/lib/api/schema')
    vi.mocked(schemaApi.listTables).mockResolvedValue({
      items: [{ name: 'orders', type: 'table', rowCount: 13 }],
    })
    useWorkspaceStore.setState({
      activeConnectionId: 'c1',
      selectedTable: 'orders',
    })

    renderWithProviders(<SchemaSubtree connectionId="c1" onSelectTable={vi.fn()} />)

    await waitFor(() => {
      expect(screen.getByText('orders')).toBeInTheDocument()
    })

    const button = screen.getByRole('button', { name: /orders/i })
    expect(button.className).toContain('bg-accent/20')
    expect(button.className).toContain('hover:bg-accent/30')
    expect(button.className).not.toContain('hover:bg-muted/40')
  })
})
