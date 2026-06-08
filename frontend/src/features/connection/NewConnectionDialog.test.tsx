import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, render, screen, fireEvent, waitFor } from '@testing-library/react'
import { NewConnectionDialog } from './NewConnectionDialog'

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    openFromFile: vi.fn(),
    createRemote: vi.fn(),
    test: vi.fn(),
  },
}))

vi.mock('@/lib/api/dialog', () => ({
  dialogApi: {
    openDatabaseFile: vi.fn(),
  },
}))

vi.mock('@tanstack/react-query', () => ({
  useMutation: ({
    mutationFn,
    onError,
    onSuccess,
  }: {
    mutationFn: (arg: unknown) => Promise<unknown>
    onError?: (err: unknown) => void
    onSuccess?: (data: unknown) => void
  }) => ({
    mutate: (arg?: unknown) => {
      void mutationFn(arg)
        .then((data) => onSuccess?.(data))
        .catch((err) => onError?.(err))
    },
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

vi.mock('@/lib/api/secrets', () => ({
  secretsApi: {
    initVault: vi.fn(),
    unlockVault: vi.fn(),
  },
  isVaultLockedError: (err: unknown) =>
    err &&
    typeof err === 'object' &&
    ((err as { code?: string }).code === 'SECRETS_VAULT_LOCKED' ||
      (err as { code?: string }).code === 'SECRETS_VAULT_NOT_INITIALIZED'),
  isVaultWrongPasswordError: () => false,
}))

vi.mock('@/stores/workspaceStore', () => ({
  useWorkspaceStore: (selector: (s: { setActiveConnectionId: () => void }) => unknown) =>
    selector({ setActiveConnectionId: vi.fn() }),
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

    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly wal={false} />)

    const input = screen.getByLabelText(/connectionForm.fields.filePath/i)
    fireEvent.change(input, { target: { value: '/tmp/test.db' } })
    fireEvent.click(screen.getByRole('button', { name: 'connectionForm.saveAndOpen' }))

    expect(connectionApi.openFromFile).toHaveBeenCalledWith({
      filePath: '/tmp/test.db',
      readOnly: true,
      wal: false,
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

    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly={false} wal />)
    fireEvent.click(screen.getByRole('button', { name: 'connectionForm.saveAndOpen' }))

    await waitFor(() => {
      expect(dialogApi.openDatabaseFile).toHaveBeenCalled()
      expect(connectionApi.openFromFile).toHaveBeenCalledWith({
        filePath: '/picked/app.db',
        readOnly: false,
        wal: true,
      })
    })
  })

  it('browse button fills path from dialog', async () => {
    const { dialogApi } = await import('@/lib/api/dialog')
    vi.mocked(dialogApi.openDatabaseFile).mockResolvedValue('/picked/browse.db')

    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly={false} wal={false} />)
    fireEvent.click(screen.getByRole('button', { name: 'connection.browse' }))

    const input = screen.getByLabelText(/connectionForm.fields.filePath/i) as HTMLInputElement
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

    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly={false} wal={false} />)
    fireEvent.click(screen.getByRole('button', { name: 'connection.browse' }))

    expect(pushToast).not.toHaveBeenCalled()
  })

  it('shows toast on browse error', async () => {
    const { dialogApi } = await import('@/lib/api/dialog')
    vi.mocked(dialogApi.openDatabaseFile).mockRejectedValue({
      code: 'INVALID_PATH',
      message: 'bad path',
    })

    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly={false} wal={false} />)
    fireEvent.click(screen.getByRole('button', { name: 'connection.browse' }))

    await waitFor(() => {
      expect(pushToast).toHaveBeenCalled()
    })
  })

  it('shows remote test when postgres dialect is selected', () => {
    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly={false} wal={false} />)
    fireEvent.click(screen.getByRole('button', { name: 'connection.type.postgres' }))
    expect(
      screen.getByRole('button', { name: 'connectionForm.testConnection' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'connectionForm.saveAndOpen' })).toBeInTheDocument()
  })

  it('shows remote form when mysql dialect is selected', () => {
    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly={false} wal={false} />)
    fireEvent.click(screen.getByRole('button', { name: 'connection.type.mysql' }))
    expect(
      screen.getByRole('button', { name: 'connectionForm.testConnection' }),
    ).toBeInTheDocument()
  })

  it('opens vault dialog and retries createRemote on vault locked', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    const { secretsApi } = await import('@/lib/api/secrets')
    vi.mocked(connectionApi.createRemote)
      .mockRejectedValueOnce({ code: 'SECRETS_VAULT_LOCKED', message: 'locked' })
      .mockResolvedValueOnce({ id: 'remote-1' } as never)
    vi.mocked(secretsApi.unlockVault).mockResolvedValue(undefined)

    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly={false} wal={false} />)
    fireEvent.click(screen.getByRole('button', { name: 'connection.type.postgres' }))
    fireEvent.change(screen.getByLabelText(/connectionForm.fields.displayName/i), {
      target: { value: 'pg-local' },
    })
    fireEvent.change(screen.getByLabelText(/connectionForm.fields.database/i), {
      target: { value: 'app' },
    })
    fireEvent.change(screen.getByLabelText(/connectionForm.fields.user/i), {
      target: { value: 'admin' },
    })
    fireEvent.change(screen.getByLabelText(/connectionForm.fields.password/i), {
      target: { value: 'secret' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'connectionForm.saveAndOpen' }))

    await waitFor(() => {
      expect(screen.getByPlaceholderText('vault.masterPassword')).toBeInTheDocument()
    })

    fireEvent.change(screen.getByPlaceholderText('vault.masterPassword'), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'vault.unlockAction' }))

    await waitFor(() => {
      expect(secretsApi.unlockVault).toHaveBeenCalledWith('password123')
      expect(connectionApi.createRemote).toHaveBeenCalledTimes(2)
    })
  })
})
