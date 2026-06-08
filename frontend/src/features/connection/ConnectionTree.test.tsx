import { describe, expect, it, vi, beforeEach } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { ConnectionTree } from './ConnectionTree'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import { useToastStore } from '@/components/ui/Toast'

vi.mock('@/features/connection/tree/ConnectionSchemaTree', () => ({
  ConnectionSchemaTree: () => <div data-testid="schema-tree" />,
}))

vi.mock('@/features/saved-queries/SavedQueries', () => ({
  SavedQueries: () => null,
}))

vi.mock('@/lib/api/connectionGroup', () => ({
  connectionGroupApi: {
    getSidebarTree: vi.fn(),
    createGroup: vi.fn(),
  },
}))

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    open: vi.fn(),
    close: vi.fn(),
    remove: vi.fn(),
  },
}))

const defaultGroup = {
  id: 'g-default',
  name: 'My Connections',
  childGroups: [],
  connections: [] as (typeof closedItem)[],
}

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

  it('shows loading then empty hint', async () => {
    const { connectionGroupApi } = await import('@/lib/api/connectionGroup')
    vi.mocked(connectionGroupApi.getSidebarTree).mockImplementation(
      () =>
        new Promise((resolve) =>
          setTimeout(() => resolve({ groups: [defaultGroup], freeConnections: [] }), 50),
        ),
    )

    renderWithProviders(<ConnectionTree />)
    expect(screen.getByText(/加载/)).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByText(/文件 → 新建连接/)).toBeInTheDocument()
    })
  })

  it('opens a closed connection on double click', async () => {
    const { connectionGroupApi } = await import('@/lib/api/connectionGroup')
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionGroupApi.getSidebarTree).mockResolvedValue({
      groups: [defaultGroup],
      freeConnections: [closedItem],
    })
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
    fireEvent.doubleClick(screen.getByText('app.db'))
    await waitFor(() => {
      expect(connectionApi.open).toHaveBeenCalledWith('c1')
    })
  })

  it('shows schema tree when connection is open', async () => {
    const { connectionGroupApi } = await import('@/lib/api/connectionGroup')
    vi.mocked(connectionGroupApi.getSidebarTree).mockResolvedValue({
      groups: [defaultGroup],
      freeConnections: [openItem],
    })

    renderWithProviders(<ConnectionTree />)
    await waitFor(() => {
      expect(screen.getByTestId('schema-tree')).toBeInTheDocument()
    })
  })

  it('creates a root group from blank area context menu', async () => {
    const { connectionGroupApi } = await import('@/lib/api/connectionGroup')
    vi.mocked(connectionGroupApi.getSidebarTree).mockResolvedValue({
      groups: [],
      freeConnections: [],
    })
    vi.mocked(connectionGroupApi.createGroup).mockResolvedValue({
      id: 'g-new',
      name: '新分组',
      sortOrder: 0,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    } as Awaited<ReturnType<typeof connectionGroupApi.createGroup>>)

    renderWithProviders(<ConnectionTree />)
    await waitFor(() => {
      expect(screen.getByText(/空白处右键/)).toBeInTheDocument()
    })

    fireEvent.contextMenu(screen.getByText(/空白处右键/))
    fireEvent.click(await screen.findByText('新建分组'))

    await waitFor(() => {
      expect(connectionGroupApi.createGroup).toHaveBeenCalledWith('', '新分组')
    })
  })
})
