import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { SqlExecutionHistoryDialog } from './SqlExecutionHistoryDialog'
import { useWorkspaceStore } from '@/stores/workspaceStore'

vi.mock('@/lib/api/queryHistory', () => ({
  queryHistoryApi: {
    listAllExecutions: vi.fn(),
  },
}))

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    list: vi.fn(),
  },
}))

describe('SqlExecutionHistoryDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useWorkspaceStore.setState({
      activeConnectionId: null,
      activeTab: 'data',
      pendingSql: null,
    })
  })

  afterEach(() => {
    cleanup()
  })

  it('renders effectRows and jumps to sql tab on row click', async () => {
    const { queryHistoryApi } = await import('@/lib/api/queryHistory')
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(queryHistoryApi.listAllExecutions).mockResolvedValue({
      items: [
        {
          id: 1,
          connectionId: 'c1',
          sql: 'SELECT 1',
          kind: 'result',
          effectRows: 42,
          durationMs: 5,
          executedAt: '2026-01-01T00:00:00Z',
        },
      ],
    })
    vi.mocked(connectionApi.list).mockResolvedValue({
      items: [{ id: 'c1', name: 'app.db', type: 'sqlite', status: 'open' }],
    } as never)

    const onOpenChange = vi.fn()
    renderWithProviders(<SqlExecutionHistoryDialog open onOpenChange={onOpenChange} />)

    await waitFor(() => {
      expect(screen.getByText('42')).toBeInTheDocument()
      expect(screen.getByText('SELECT 1')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('SELECT 1'))

    expect(useWorkspaceStore.getState().activeConnectionId).toBe('c1')
    expect(useWorkspaceStore.getState().activeTab).toBe('sql')
    expect(useWorkspaceStore.getState().pendingSql).toBe('SELECT 1')
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })
})
