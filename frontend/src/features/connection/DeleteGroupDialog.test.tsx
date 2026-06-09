import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { DeleteGroupDialog } from './DeleteGroupDialog'
import '@/i18n'

vi.mock('@/lib/api/connectionGroup', () => ({
  connectionGroupApi: {
    getGroupDeletePreview: vi.fn(),
    deleteGroup: vi.fn(),
  },
}))

vi.mock('@/components/ui/Toast', () => ({
  useToastStore: () => vi.fn(),
}))

function renderDialog(
  props: React.ComponentProps<typeof DeleteGroupDialog>,
  qc = new QueryClient({ defaultOptions: { queries: { retry: false } } }),
) {
  return {
    qc,
    ...render(
      <QueryClientProvider client={qc}>
        <DeleteGroupDialog {...props} />
      </QueryClientProvider>,
    ),
  }
}

describe('DeleteGroupDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('shows simple confirm when group is empty', async () => {
    const { connectionGroupApi } = await import('@/lib/api/connectionGroup')
    vi.mocked(connectionGroupApi.getGroupDeletePreview).mockResolvedValue({
      subgroupCount: 0,
      connCount: 0,
    })
    vi.mocked(connectionGroupApi.deleteGroup).mockResolvedValue(undefined)

    const onDeleted = vi.fn()
    renderDialog({
      groupId: 'g1',
      groupName: 'Work',
      open: true,
      onOpenChange: () => {},
      onDeleted,
    })

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

  it('shows checkbox and subgroup count when group has connections', async () => {
    const { connectionGroupApi } = await import('@/lib/api/connectionGroup')
    vi.mocked(connectionGroupApi.getGroupDeletePreview).mockResolvedValue({
      subgroupCount: 2,
      connCount: 2,
    })
    vi.mocked(connectionGroupApi.deleteGroup).mockResolvedValue(undefined)

    renderDialog({
      groupId: 'g1',
      groupName: 'Work',
      open: true,
      onOpenChange: () => {},
      onDeleted: () => {},
    })

    await waitFor(() => {
      expect(screen.getByRole('checkbox')).toBeInTheDocument()
      expect(screen.getByText(/2.*子分组/)).toBeInTheDocument()
    })
  })

  it('resets deleteConnections checkbox when dialog closes', async () => {
    const { connectionGroupApi } = await import('@/lib/api/connectionGroup')
    vi.mocked(connectionGroupApi.getGroupDeletePreview).mockResolvedValue({
      subgroupCount: 0,
      connCount: 2,
    })

    const onOpenChange = vi.fn()
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    const { rerender } = render(
      <QueryClientProvider client={qc}>
        <DeleteGroupDialog
          groupId="g1"
          groupName="Work"
          open
          onOpenChange={onOpenChange}
          onDeleted={() => {}}
        />
      </QueryClientProvider>,
    )

    await waitFor(() => expect(screen.getByRole('checkbox')).toBeInTheDocument())
    fireEvent.click(screen.getByRole('checkbox'))
    fireEvent.click(screen.getByRole('button', { name: /cancel|取消/i }))

    expect(onOpenChange).toHaveBeenCalledWith(false)

    rerender(
      <QueryClientProvider client={qc}>
        <DeleteGroupDialog
          groupId="g2"
          groupName="Other"
          open
          onOpenChange={onOpenChange}
          onDeleted={() => {}}
        />
      </QueryClientProvider>,
    )

    await waitFor(() => {
      const checkbox = screen.getByRole('checkbox') as HTMLInputElement
      expect(checkbox.checked).toBe(false)
    })
  })
})
