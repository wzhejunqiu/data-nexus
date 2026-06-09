import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { AttachDatabaseDialog } from './AttachDatabaseDialog'

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    attach: vi.fn(),
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
})
