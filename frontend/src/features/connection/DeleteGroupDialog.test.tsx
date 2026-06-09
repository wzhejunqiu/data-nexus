import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { DeleteGroupDialog } from './DeleteGroupDialog'

vi.mock('@/lib/api/connectionGroup', () => ({
  connectionGroupApi: {
    countConnectionsInGroup: vi.fn(),
    deleteGroup: vi.fn(),
  },
}))

vi.mock('@/components/ui/Toast', () => ({
  useToastStore: () => vi.fn(),
}))

describe('DeleteGroupDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('shows simple confirm when group is empty', async () => {
    const { connectionGroupApi } = await import('@/lib/api/connectionGroup')
    vi.mocked(connectionGroupApi.countConnectionsInGroup).mockResolvedValue(0)
    vi.mocked(connectionGroupApi.deleteGroup).mockResolvedValue(undefined)

    const onDeleted = vi.fn()
    renderWithProviders(
      <DeleteGroupDialog
        groupId="g1"
        groupName="Work"
        open
        onOpenChange={() => {}}
        onDeleted={onDeleted}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText(/Work/)).toBeInTheDocument()
    })
    expect(screen.queryByRole('checkbox')).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: '删除分组' }))
    await waitFor(() => {
      expect(connectionGroupApi.deleteGroup).toHaveBeenCalledWith({
        id: 'g1',
        deleteConnections: false,
      })
      expect(onDeleted).toHaveBeenCalled()
    })
  })

  it('shows checkbox when group has connections', async () => {
    const { connectionGroupApi } = await import('@/lib/api/connectionGroup')
    vi.mocked(connectionGroupApi.countConnectionsInGroup).mockResolvedValue(2)
    vi.mocked(connectionGroupApi.deleteGroup).mockResolvedValue(undefined)

    renderWithProviders(
      <DeleteGroupDialog
        groupId="g1"
        groupName="Work"
        open
        onOpenChange={() => {}}
        onDeleted={() => {}}
      />,
    )

    await waitFor(() => {
      expect(screen.getByRole('checkbox')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('checkbox'))
    fireEvent.click(screen.getByRole('button', { name: '删除分组' }))

    await waitFor(() => {
      expect(connectionGroupApi.deleteGroup).toHaveBeenCalledWith({
        id: 'g1',
        deleteConnections: true,
      })
    })
  })
})
