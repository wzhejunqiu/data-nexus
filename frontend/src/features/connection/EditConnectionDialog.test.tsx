import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { EditConnectionDialog } from './EditConnectionDialog'
import type { ConnectionListItem } from '@/lib/types'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    updateSQLiteSettings: vi.fn(),
    updatePostgresSettings: vi.fn(),
    updateMySQLSettings: vi.fn(),
    rename: vi.fn(),
  },
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

vi.mock('@/components/ui/Toast', () => ({
  useToastStore: () => vi.fn(),
}))

const baseItem: ConnectionListItem = {
  id: 'conn-1',
  name: 'test.db',
  type: 'sqlite',
  config: {
    type: 'sqlite',
    sqlite: { filePath: '/tmp/test.db', readOnly: false, wal: false },
  },
  status: 'closed',
  lastUsedAt: '2026-01-01T00:00:00Z',
}

const postgresItem: ConnectionListItem = {
  id: 'pg-1',
  name: 'local-pg',
  type: 'postgres',
  config: {
    type: 'postgres',
    postgres: {
      host: 'localhost',
      port: 5432,
      database: 'app',
      user: 'admin',
      sslMode: 'disable',
      schema: 'public',
      clientEncoding: 'LATIN1',
      readOnly: false,
    },
  },
  status: 'closed',
  lastUsedAt: '2026-01-01T00:00:00Z',
}

function renderDialog(item: ConnectionListItem = baseItem) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <EditConnectionDialog open item={item} onOpenChange={() => {}} onSaved={() => {}} />
    </QueryClientProvider>,
  )
}

describe('EditConnectionDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders readOnly and wal checkboxes with initial values', () => {
    renderDialog()
    fireEvent.click(screen.getByRole('button', { name: /connectionForm.section.advanced/ }))
    const checkboxes = screen.getAllByRole('checkbox') as HTMLInputElement[]
    expect(checkboxes[0].checked).toBe(false)
    expect(checkboxes[1].checked).toBe(false)
  })

  it('shows readOnly checked when connection is read-only', () => {
    renderDialog({
      ...baseItem,
      config: {
        type: 'sqlite',
        sqlite: { filePath: '/tmp/test.db', readOnly: true, wal: false },
      },
    })
    fireEvent.click(screen.getByRole('button', { name: /connectionForm.section.advanced/ }))
    const checkboxes = screen.getAllByRole('checkbox') as HTMLInputElement[]
    expect(checkboxes[0].checked).toBe(true)
    expect(checkboxes[1].disabled).toBe(true)
  })

  it('calls updateSQLiteSettings when saving toggled readOnly', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.updateSQLiteSettings).mockResolvedValue({} as never)

    renderDialog()
    fireEvent.click(screen.getByRole('button', { name: /connectionForm.section.advanced/ }))
    const checkboxes = screen.getAllByRole('checkbox')
    fireEvent.click(checkboxes[0])
    fireEvent.click(screen.getByRole('button', { name: /confirm|确认/i }))

    await waitFor(() => {
      expect(connectionApi.updateSQLiteSettings).toHaveBeenCalledWith('conn-1', {
        readOnly: true,
        wal: false,
      })
    })
  })

  it('calls updateSQLiteSettings with wal enabled', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.updateSQLiteSettings).mockResolvedValue({} as never)

    renderDialog()
    fireEvent.click(screen.getByRole('button', { name: /connectionForm.section.advanced/ }))
    const checkboxes = screen.getAllByRole('checkbox')
    fireEvent.click(checkboxes[1])
    fireEvent.click(screen.getByRole('button', { name: /confirm|确认/i }))

    await waitFor(() => {
      expect(connectionApi.updateSQLiteSettings).toHaveBeenCalledWith('conn-1', {
        readOnly: false,
        wal: true,
      })
    })
  })

  it('saves postgres settings without password when left empty', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.updatePostgresSettings).mockResolvedValue({} as never)

    renderDialog(postgresItem)
    fireEvent.click(screen.getByRole('button', { name: /confirm|确认/i }))

    await waitFor(() => {
      expect(connectionApi.updatePostgresSettings).toHaveBeenCalledWith('pg-1', {
        host: 'localhost',
        port: 5432,
        database: 'app',
        user: 'admin',
        sslMode: 'disable',
        schema: 'public',
        clientEncoding: 'LATIN1',
        readOnly: false,
        password: undefined,
      })
    })
  })

  it('saves mysql advanced settings including charset and storage engine', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.updateMySQLSettings).mockResolvedValue({} as never)

    const mysqlItem: ConnectionListItem = {
      id: 'my-1',
      name: 'local-mysql',
      type: 'mysql',
      config: {
        type: 'mysql',
        mysql: {
          host: 'localhost',
          port: 3306,
          database: 'app',
          user: 'root',
          tls: false,
          charset: 'gbk',
          collation: 'gbk_chinese_ci',
          defaultStorageEngine: 'InnoDB',
          readOnly: false,
        },
      },
      status: 'closed',
      lastUsedAt: '2026-01-01T00:00:00Z',
    }

    renderDialog(mysqlItem)
    fireEvent.click(screen.getByRole('button', { name: /confirm|确认/i }))

    await waitFor(() => {
      expect(connectionApi.updateMySQLSettings).toHaveBeenCalledWith('my-1', {
        host: 'localhost',
        port: 3306,
        database: 'app',
        user: 'root',
        tls: false,
        tlsSkipVerify: false,
        charset: 'gbk',
        collation: 'gbk_chinese_ci',
        defaultStorageEngine: 'InnoDB',
        readOnly: false,
        password: undefined,
      })
    })
  })

  it('opens vault dialog and retries save when vault is locked', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    const { secretsApi } = await import('@/lib/api/secrets')
    vi.mocked(connectionApi.updatePostgresSettings)
      .mockRejectedValueOnce({ code: 'SECRETS_VAULT_LOCKED', message: 'locked' })
      .mockResolvedValueOnce({} as never)
    vi.mocked(secretsApi.unlockVault).mockResolvedValue(undefined)

    renderDialog(postgresItem)
    fireEvent.click(screen.getByRole('button', { name: /confirm|确认/i }))

    await waitFor(() => {
      expect(screen.getByPlaceholderText('vault.masterPassword')).toBeInTheDocument()
    })

    fireEvent.change(screen.getByPlaceholderText('vault.masterPassword'), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'vault.unlockAction' }))

    await waitFor(() => {
      expect(secretsApi.unlockVault).toHaveBeenCalledWith('password123')
      expect(connectionApi.updatePostgresSettings).toHaveBeenCalledTimes(2)
    })
  })
})
