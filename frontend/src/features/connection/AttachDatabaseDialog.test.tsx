import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { AttachDatabaseDialog } from './AttachDatabaseDialog'

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    attach: vi.fn(),
    listAttached: vi.fn().mockResolvedValue([]),
  },
}))

vi.mock('@/lib/api/dialog', () => ({
  dialogApi: {
    openDatabaseFile: vi.fn(),
  },
}))

vi.mock('@/components/ui/Toast', () => ({
  useToastStore: () => vi.fn(),
}))

describe('AttachDatabaseDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('prefills path from initialPath and submits attach', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.attach).mockResolvedValue(undefined)

    const onAttached = vi.fn()
    const onOpenChange = vi.fn()
    renderWithProviders(
      <AttachDatabaseDialog
        connectionId="c1"
        open
        initialPath="/tmp/extra.db"
        onOpenChange={onOpenChange}
        onAttached={onAttached}
      />,
    )

    const pathInput = screen.getAllByRole('textbox')[0] as HTMLInputElement
    expect(pathInput.value).toBe('/tmp/extra.db')

    fireEvent.click(screen.getByRole('button', { name: '附加' }))

    await waitFor(() => {
      expect(connectionApi.attach).toHaveBeenCalledWith('c1', '/tmp/extra.db', 'extra')
      expect(onAttached).toHaveBeenCalled()
      expect(onOpenChange).toHaveBeenCalledWith(false)
    })
  })

  it('shows session hint', () => {
    renderWithProviders(
      <AttachDatabaseDialog connectionId="c1" open onOpenChange={vi.fn()} onAttached={vi.fn()} />,
    )

    expect(screen.getByText('附加为当前会话临时挂载，关闭连接后失效')).toBeInTheDocument()
    expect(screen.queryByText('附加库将以只读模式挂载')).not.toBeInTheDocument()
  })

  it('shows read-only hint when connection is read-only', () => {
    renderWithProviders(
      <AttachDatabaseDialog
        connectionId="c1"
        open
        readOnly
        onOpenChange={vi.fn()}
        onAttached={vi.fn()}
      />,
    )

    expect(screen.getByText('附加库将以只读模式挂载')).toBeInTheDocument()
  })

  it('disables attach when alias is reserved', async () => {
    renderWithProviders(
      <AttachDatabaseDialog
        connectionId="c1"
        open
        initialPath="/tmp/extra.db"
        onOpenChange={vi.fn()}
        onAttached={vi.fn()}
      />,
    )

    const aliasInput = screen.getAllByRole('textbox')[1] as HTMLInputElement
    fireEvent.change(aliasInput, { target: { value: 'main' } })

    expect(screen.getByText('别名不能为 main 或 temp')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '附加' })).toBeDisabled()
  })
})
