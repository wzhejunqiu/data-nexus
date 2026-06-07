import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, render, screen, fireEvent, waitFor } from '@testing-library/react'
import { NewConnectionDialog } from './NewConnectionDialog'

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    openFromFile: vi.fn(),
  },
}))

vi.mock('@/lib/api/dialog', () => ({
  dialogApi: {
    openDatabaseFile: vi.fn(),
  },
}))

vi.mock('@tanstack/react-query', () => ({
  useMutation: ({ mutationFn }: { mutationFn: () => Promise<unknown> }) => ({
    mutate: () => mutationFn(),
    isPending: false,
  }),
  useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

const pushToast = vi.fn()
vi.mock('@/components/ui/Toast', () => ({
  useToastStore: (selector: (s: { push: typeof pushToast }) => unknown) =>
    selector({ push: pushToast }),
}))

describe('NewConnectionDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    pushToast.mockClear()
  })

  afterEach(() => {
    cleanup()
  })

  it('passes readOnly flag when opening from path', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.openFromFile).mockResolvedValue({
      id: '1',
      displayName: 'test.db',
    } as never)

    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly />)

    const input = screen.getByPlaceholderText('/path/to/database.db')
    fireEvent.change(input, { target: { value: '/tmp/test.db' } })
    fireEvent.click(screen.getByRole('button', { name: 'connection.open' }))

    expect(connectionApi.openFromFile).toHaveBeenCalledWith({
      filePath: '/tmp/test.db',
      readOnly: true,
    })
  })

  it('opens from file dialog when path is empty', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    const { dialogApi } = await import('@/lib/api/dialog')
    vi.mocked(dialogApi.openDatabaseFile).mockResolvedValue('/picked/app.db')
    vi.mocked(connectionApi.openFromFile).mockResolvedValue({
      id: '1',
      displayName: 'app.db',
    } as never)

    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly={false} />)
    fireEvent.click(screen.getByRole('button', { name: 'connection.open' }))

    await waitFor(() => {
      expect(dialogApi.openDatabaseFile).toHaveBeenCalled()
      expect(connectionApi.openFromFile).toHaveBeenCalledWith({
        filePath: '/picked/app.db',
        readOnly: false,
      })
    })
  })

  it('browse button fills path from dialog', async () => {
    const { dialogApi } = await import('@/lib/api/dialog')
    vi.mocked(dialogApi.openDatabaseFile).mockResolvedValue('/picked/browse.db')

    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly={false} />)
    fireEvent.click(screen.getByRole('button', { name: 'connection.browse' }))

    const input = screen.getByPlaceholderText('/path/to/database.db') as HTMLInputElement
    await waitFor(() => {
      expect(input.value).toBe('/picked/browse.db')
    })
  })

  it('ignores dialog cancelled on browse', async () => {
    const { dialogApi } = await import('@/lib/api/dialog')
    vi.mocked(dialogApi.openDatabaseFile).mockRejectedValue({
      code: 'DIALOG_CANCELLED',
      message: 'cancelled',
    })

    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly={false} />)
    fireEvent.click(screen.getByRole('button', { name: 'connection.browse' }))

    expect(pushToast).not.toHaveBeenCalled()
  })

  it('shows toast on browse error', async () => {
    const { dialogApi } = await import('@/lib/api/dialog')
    vi.mocked(dialogApi.openDatabaseFile).mockRejectedValue({
      code: 'INVALID_PATH',
      message: 'bad path',
    })

    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly={false} />)
    fireEvent.click(screen.getByRole('button', { name: 'connection.browse' }))

    await waitFor(() => {
      expect(pushToast).toHaveBeenCalled()
    })
  })
})
