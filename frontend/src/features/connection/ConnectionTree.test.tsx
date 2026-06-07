import { describe, expect, it, vi, beforeEach } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { ConnectionTree } from './ConnectionTree'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import { useToastStore } from '@/components/ui/Toast'

vi.mock('@/features/schema/SchemaSubtree', () => ({
  SchemaSubtree: () => <div data-testid="schema-subtree" />,
}))

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    list: vi.fn(),
    open: vi.fn(),
    close: vi.fn(),
    remove: vi.fn(),
    rename: vi.fn(),
    getRestoreOpenOnStartup: vi.fn(),
    setRestoreOpenOnStartup: vi.fn(),
  },
}))

const closedItem = {
  id: 'c1',
  name: 'app.db',
  type: 'sqlite' as const,
  config: { type: 'sqlite' as const, sqlite: { filePath: '/tmp/app.db', readOnly: false } },
  status: 'closed' as const,
  lastUsedAt: '2026-01-01T00:00:00Z',
}

const openItem = { ...closedItem, id: 'c2', status: 'open' as const }

describe('ConnectionTree', () => {
  beforeEach(() => {
    useWorkspaceStore.setState({
      activeConnectionId: null,
      selectedTable: null,
      activeTab: 'data',
    })
    useToastStore.setState({ items: [] })
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })

  it('shows loading then empty state', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.list).mockImplementation(
      () => new Promise((resolve) => setTimeout(() => resolve({ items: [] }), 50)),
    )
    vi.mocked(connectionApi.getRestoreOpenOnStartup).mockResolvedValue(false)

    renderWithProviders(<ConnectionTree />)
    expect(screen.getByText(/加载/)).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByText(/新建或打开一个连接开始/)).toBeInTheDocument()
    })
  })

  it('opens a closed connection', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [closedItem] })
    vi.mocked(connectionApi.getRestoreOpenOnStartup).mockResolvedValue(false)
    vi.mocked(connectionApi.open).mockResolvedValue({
      id: 'c1',
      type: 'sqlite',
      displayName: 'app.db',
      config: closedItem.config,
      connectedAt: '2026-01-01T00:00:00Z',
    })

    renderWithProviders(<ConnectionTree />)
    await waitFor(() => {
      expect(screen.getByText('app.db')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: '打开' }))
    await waitFor(() => {
      expect(connectionApi.open).toHaveBeenCalledWith('c1')
    })
  })

  it('closes an open connection', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [openItem] })
    vi.mocked(connectionApi.getRestoreOpenOnStartup).mockResolvedValue(false)
    vi.mocked(connectionApi.close).mockResolvedValue(undefined)

    renderWithProviders(<ConnectionTree />)
    await waitFor(() => {
      expect(screen.getByTestId('schema-subtree')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: '关闭' }))
    await waitFor(() => {
      expect(connectionApi.close).toHaveBeenCalledWith('c2')
    })
  })

  it('does not remove when confirm is cancelled', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [closedItem] })
    vi.mocked(connectionApi.getRestoreOpenOnStartup).mockResolvedValue(false)
    vi.spyOn(window, 'confirm').mockReturnValue(false)

    renderWithProviders(<ConnectionTree />)
    await waitFor(() => {
      expect(screen.getByText('app.db')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: '删除' }))
    expect(connectionApi.remove).not.toHaveBeenCalled()
  })

  it('removes when confirm is accepted', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [closedItem] })
    vi.mocked(connectionApi.getRestoreOpenOnStartup).mockResolvedValue(false)
    vi.mocked(connectionApi.remove).mockResolvedValue(undefined)

    renderWithProviders(<ConnectionTree />)
    await waitFor(() => {
      expect(screen.getByText('app.db')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: '删除' }))
    await waitFor(() => {
      expect(connectionApi.remove).toHaveBeenCalledWith('c1')
    })
  })

  it('toggles restore on startup', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [] })
    vi.mocked(connectionApi.getRestoreOpenOnStartup).mockResolvedValue(false)
    vi.mocked(connectionApi.setRestoreOpenOnStartup).mockResolvedValue(undefined)

    renderWithProviders(<ConnectionTree />)
    const checkbox = await screen.findByLabelText(/启动时恢复已打开连接/)
    fireEvent.click(checkbox)
    await waitFor(() => {
      expect(connectionApi.setRestoreOpenOnStartup).toHaveBeenCalledWith(true)
    })
  })

  it('opens new connection dialog', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [] })
    vi.mocked(connectionApi.getRestoreOpenOnStartup).mockResolvedValue(false)

    renderWithProviders(<ConnectionTree />)
    fireEvent.click(screen.getByRole('button', { name: '新建连接' }))
    expect(screen.getByPlaceholderText('/path/to/database.db')).toBeInTheDocument()
  })
})
