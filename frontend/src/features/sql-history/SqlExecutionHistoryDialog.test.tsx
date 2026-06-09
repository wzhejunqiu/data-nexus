import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { SqlExecutionHistoryDialog } from './SqlExecutionHistoryDialog'
import { useWorkspaceStore } from '@/stores/workspaceStore'

const pushToast = vi.fn()

vi.mock('@/lib/api/queryHistory', () => ({
  queryHistoryApi: {
    listAllExecutions: vi.fn(),
  },
}))

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    list: vi.fn(),
    open: vi.fn(),
  },
}))

vi.mock('@/components/ui/Toast', () => ({
  useToastStore: (selector: (s: { push: typeof pushToast }) => unknown) =>
    selector({ push: pushToast }),
}))

const historyRow = {
  id: 1,
  connectionId: 'c1',
  sql: 'SELECT 1',
  kind: 'result' as const,
  effectRows: 42,
  durationMs: 5,
  executedAt: '2026-01-01T00:00:00Z',
}

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

  async function setupHistory(status: 'open' | 'closed' = 'open') {
    const { queryHistoryApi } = await import('@/lib/api/queryHistory')
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(queryHistoryApi.listAllExecutions).mockResolvedValue({ items: [historyRow] })
    vi.mocked(connectionApi.list).mockResolvedValue({
      items: [{ id: 'c1', name: 'app.db', type: 'sqlite', status }],
    } as never)
    vi.mocked(connectionApi.open).mockResolvedValue({ id: 'c1' } as never)
  }

  it('renders effectRows and jumps to sql tab when connection is open', async () => {
    await setupHistory('open')
    const { connectionApi } = await import('@/lib/api/connection')
    const onOpenChange = vi.fn()
    renderWithProviders(<SqlExecutionHistoryDialog open onOpenChange={onOpenChange} />)

    await waitFor(() => {
      expect(screen.getByText('42')).toBeInTheDocument()
      expect(screen.getByText('SELECT 1')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('SELECT 1'))

    await waitFor(() => {
      expect(connectionApi.open).not.toHaveBeenCalled()
      expect(useWorkspaceStore.getState().activeConnectionId).toBe('c1')
      expect(useWorkspaceStore.getState().activeTab).toBe('sql')
      expect(useWorkspaceStore.getState().pendingSql).toBe('SELECT 1')
      expect(onOpenChange).toHaveBeenCalledWith(false)
    })
  })

  it('opens closed connection before jumping', async () => {
    await setupHistory('closed')
    const { connectionApi } = await import('@/lib/api/connection')
    const onOpenChange = vi.fn()
    renderWithProviders(<SqlExecutionHistoryDialog open onOpenChange={onOpenChange} />)

    await waitFor(() => expect(screen.getByText('SELECT 1')).toBeInTheDocument())
    fireEvent.click(screen.getByText('SELECT 1'))

    await waitFor(() => {
      expect(connectionApi.open).toHaveBeenCalledWith('c1')
      expect(useWorkspaceStore.getState().activeConnectionId).toBe('c1')
      expect(onOpenChange).toHaveBeenCalledWith(false)
    })
  })

  it('shows toast and does not jump when connection is missing', async () => {
    const { queryHistoryApi } = await import('@/lib/api/queryHistory')
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(queryHistoryApi.listAllExecutions).mockResolvedValue({ items: [historyRow] })
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [] } as never)

    const onOpenChange = vi.fn()
    renderWithProviders(<SqlExecutionHistoryDialog open onOpenChange={onOpenChange} />)

    await waitFor(() => expect(screen.getByText('SELECT 1')).toBeInTheDocument())
    fireEvent.click(screen.getByText('SELECT 1'))

    await waitFor(() => {
      expect(pushToast).toHaveBeenCalled()
      expect(useWorkspaceStore.getState().activeConnectionId).toBeNull()
      expect(onOpenChange).not.toHaveBeenCalled()
    })
  })

  it('shows deleted label for missing connection in list', async () => {
    const { queryHistoryApi } = await import('@/lib/api/queryHistory')
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(queryHistoryApi.listAllExecutions).mockResolvedValue({ items: [historyRow] })
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [] } as never)

    renderWithProviders(<SqlExecutionHistoryDialog open onOpenChange={() => {}} />)

    await waitFor(() => {
      expect(screen.getByText('（已删除）')).toBeInTheDocument()
    })
  })
})
