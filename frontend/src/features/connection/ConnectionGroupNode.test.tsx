import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react'
import { DndContext } from '@dnd-kit/core'
import { renderWithProviders } from '@/test/render'
import { ConnectionGroupNode } from './ConnectionGroupNode'
import { invokeSidebarRename } from './sidebarRenameHandlers'
import { useWorkspaceStore } from '@/stores/workspaceStore'

vi.mock('@/lib/api/connectionGroup', () => ({
  connectionGroupApi: {
    renameGroup: vi.fn(),
    createGroup: vi.fn(),
    getGroupDeletePreview: vi.fn().mockResolvedValue({ subgroupCount: 0, connCount: 0 }),
    deleteGroup: vi.fn(),
  },
}))

vi.mock('@/components/ui/Toast', () => ({
  useToastStore: () => vi.fn(),
}))

vi.mock('./ConnectionTreeItem', () => ({
  ConnectionTreeItem: () => null,
}))

vi.mock('./dnd/SidebarDndContext', () => ({
  GroupSortableList: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const node = {
  id: 'g1',
  name: 'Work',
  childGroups: [],
  connections: [],
}

function renderGroup() {
  return renderWithProviders(
    <DndContext onDragEnd={() => {}}>
      <ConnectionGroupNode node={node} depth={0} sortContainerId="root" onRefresh={() => {}} />
    </DndContext>,
  )
}

describe('ConnectionGroupNode', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useWorkspaceStore.setState({ sidebarFocus: { kind: 'group', id: 'g1' } })
  })

  afterEach(() => {
    cleanup()
  })

  it('enters inline rename on F2 handler and commits on Enter', async () => {
    const { connectionGroupApi } = await import('@/lib/api/connectionGroup')
    vi.mocked(connectionGroupApi.renameGroup).mockResolvedValue({
      id: 'g1',
      name: 'Projects',
      sortOrder: 0,
    } as never)

    renderGroup()
    invokeSidebarRename({ kind: 'group', id: 'g1' })

    const input = await screen.findByDisplayValue('Work')
    fireEvent.change(input, { target: { value: 'Projects' } })
    fireEvent.keyDown(input, { key: 'Enter' })

    await waitFor(() => {
      expect(connectionGroupApi.renameGroup).toHaveBeenCalledWith('g1', 'Projects')
    })
  })

  it('does not call rename when name is empty', async () => {
    const { connectionGroupApi } = await import('@/lib/api/connectionGroup')
    renderGroup()
    invokeSidebarRename({ kind: 'group', id: 'g1' })

    const input = await screen.findByDisplayValue('Work')
    fireEvent.change(input, { target: { value: '   ' } })
    fireEvent.keyDown(input, { key: 'Enter' })

    expect(connectionGroupApi.renameGroup).not.toHaveBeenCalled()
  })
})
