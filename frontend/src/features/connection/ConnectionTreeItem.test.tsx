import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react'
import { DndContext } from '@dnd-kit/core'
import { renderWithProviders } from '@/test/render'
import { ConnectionTreeItem } from './ConnectionTreeItem'
import { invokeSidebarRename } from './sidebarRenameHandlers'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import type { ConnectionListItem } from '@/lib/types'

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    rename: vi.fn(),
    open: vi.fn(),
    close: vi.fn(),
    remove: vi.fn(),
  },
}))

vi.mock('@/components/ui/Toast', () => ({
  useToastStore: () => vi.fn(),
}))

vi.mock('./tree/ConnectionSchemaTree', () => ({
  ConnectionSchemaTree: () => null,
}))

vi.mock('./EditConnectionDialog', () => ({
  EditConnectionDialog: () => null,
}))

const closedItem: ConnectionListItem = {
  id: 'c1',
  name: 'app.db',
  type: 'sqlite',
  config: {
    type: 'sqlite',
    sqlite: { filePath: '/tmp/app.db', readOnly: false, wal: false },
  },
  status: 'closed',
  lastUsedAt: '2026-01-01T00:00:00Z',
}

const openItem: ConnectionListItem = { ...closedItem, id: 'c2', status: 'open' }

function renderItem(item: ConnectionListItem) {
  return renderWithProviders(
    <DndContext onDragEnd={() => {}}>
      <ConnectionTreeItem item={item} depth={0} sortContainerId="g1" onRefresh={() => {}} />
    </DndContext>,
  )
}

describe('ConnectionTreeItem', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it.each([
    ['closed', closedItem],
    ['open', openItem],
  ])('inline renames %s connection on Enter', async (_label, item) => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.rename).mockResolvedValue({
      id: item.id,
      name: 'renamed.db',
      type: item.type,
      config: item.config,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
      lastUsedAt: '2026-01-01T00:00:00Z',
    } as never)
    useWorkspaceStore.setState({ sidebarFocus: { kind: 'connection', id: item.id } })

    renderItem(item)
    invokeSidebarRename({ kind: 'connection', id: item.id })

    const input = await screen.findByDisplayValue(item.name)
    fireEvent.change(input, { target: { value: 'renamed.db' } })
    fireEvent.keyDown(input, { key: 'Enter' })

    await waitFor(() => {
      expect(connectionApi.rename).toHaveBeenCalledWith(item.id, 'renamed.db')
    })
  })

  it('does not call rename for empty name', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    useWorkspaceStore.setState({ sidebarFocus: { kind: 'connection', id: closedItem.id } })

    renderItem(closedItem)
    invokeSidebarRename({ kind: 'connection', id: closedItem.id })

    const input = await screen.findByDisplayValue('app.db')
    fireEvent.change(input, { target: { value: '' } })
    fireEvent.keyDown(input, { key: 'Enter' })

    expect(connectionApi.rename).not.toHaveBeenCalled()
  })

  it('calls onOpenAttach when attach menu is selected on open sqlite', async () => {
    const onOpenAttach = vi.fn()
    renderWithProviders(
      <DndContext onDragEnd={() => {}}>
        <ConnectionTreeItem
          item={openItem}
          depth={0}
          sortContainerId="g1"
          onRefresh={() => {}}
          onOpenAttach={onOpenAttach}
        />
      </DndContext>,
    )

    fireEvent.contextMenu(screen.getByText('app.db'))
    fireEvent.click(await screen.findByText('附加数据库…'))

    expect(onOpenAttach).toHaveBeenCalledWith({ connectionId: 'c2' })
  })
})
