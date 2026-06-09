import { describe, expect, it, vi, beforeEach } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { ConnectionSchemaTree } from './ConnectionSchemaTree'
import type { ConnectionListItem } from '@/lib/types'

const listNamespaces = vi.fn()
const listSchemas = vi.fn()
const listTables = vi.fn()

vi.mock('@/lib/api/schema', () => ({
  schemaApi: {
    listNamespaces: (...args: unknown[]) => listNamespaces(...args),
    listSchemas: (...args: unknown[]) => listSchemas(...args),
    listTables: (...args: unknown[]) => listTables(...args),
  },
}))

vi.mock('@/lib/api/app', () => ({
  appApi: {
    getPlatform: vi.fn().mockResolvedValue('darwin'),
    revealFileInExplorer: vi.fn(),
  },
}))

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    detach: vi.fn(),
  },
}))

function baseItem(type: 'sqlite' | 'mysql' | 'postgres'): ConnectionListItem {
  return {
    id: 'conn-1',
    name: 'db',
    type,
    config:
      type === 'sqlite'
        ? { type: 'sqlite', sqlite: { filePath: '/tmp/app.db', readOnly: false } }
        : type === 'mysql'
          ? {
              type: 'mysql',
              mysql: {
                host: 'localhost',
                port: 3306,
                database: 'app',
                user: 'root',
                tls: false,
                readOnly: false,
              },
            }
          : {
              type: 'postgres',
              postgres: {
                host: 'localhost',
                port: 5432,
                database: 'app',
                user: 'admin',
                sslMode: 'disable',
                schema: 'public',
                readOnly: false,
              },
            },
    status: 'open',
    lastUsedAt: '2026-01-01T00:00:00Z',
  }
}

describe('ConnectionSchemaTree lazy loading', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listNamespaces.mockResolvedValue({ items: [] })
    listSchemas.mockResolvedValue({ items: [] })
    listTables.mockResolvedValue({ items: [{ name: 'users', type: 'table' }] })
  })

  it('loads namespaces on mount only for sqlite', async () => {
    listNamespaces.mockResolvedValue({ items: [{ name: 'main', kind: 'main' }] })

    renderWithProviders(<ConnectionSchemaTree item={baseItem('sqlite')} onSelectTable={() => {}} />)

    await waitFor(() => expect(listNamespaces).toHaveBeenCalledWith('conn-1'))
    expect(listTables).not.toHaveBeenCalled()
    expect(listSchemas).not.toHaveBeenCalled()
  })

  it('loads tables after expanding sqlite namespace', async () => {
    listNamespaces.mockResolvedValue({ items: [{ name: 'main', kind: 'main' }] })

    renderWithProviders(<ConnectionSchemaTree item={baseItem('sqlite')} onSelectTable={() => {}} />)

    await waitFor(() => expect(screen.getByText('main')).toBeInTheDocument())
    fireEvent.click(screen.getByText('main'))

    await waitFor(() =>
      expect(listTables).toHaveBeenCalledWith('conn-1', { database: 'main', schema: undefined }),
    )
    expect(listSchemas).not.toHaveBeenCalled()
  })

  it('loads tables after expanding mysql namespace', async () => {
    listNamespaces.mockResolvedValue({ items: [{ name: 'app', kind: 'database' }] })

    renderWithProviders(<ConnectionSchemaTree item={baseItem('mysql')} onSelectTable={() => {}} />)

    await waitFor(() => expect(screen.getByText('app')).toBeInTheDocument())
    fireEvent.click(screen.getByText('app'))

    await waitFor(() =>
      expect(listTables).toHaveBeenCalledWith('conn-1', { database: 'app', schema: undefined }),
    )
    expect(listSchemas).not.toHaveBeenCalled()
  })

  it('selects mysql table with database-qualified name', async () => {
    listNamespaces.mockResolvedValue({ items: [{ name: 'app', kind: 'database' }] })
    listTables.mockResolvedValue({
      items: [{ name: 'users', type: 'table', schema: 'app' }],
    })
    const onSelectTable = vi.fn()

    renderWithProviders(
      <ConnectionSchemaTree item={baseItem('mysql')} onSelectTable={onSelectTable} />,
    )

    await waitFor(() => expect(screen.getByText('app')).toBeInTheDocument())
    fireEvent.click(screen.getByText('app'))
    await waitFor(() => expect(screen.getByRole('button', { name: /users/ })).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: /users/ }))

    expect(onSelectTable).toHaveBeenCalledWith('app.users', { database: 'app', schema: undefined })
  })

  it('loads schemas then tables for postgres', async () => {
    listNamespaces.mockResolvedValue({ items: [{ name: 'appdb', kind: 'database' }] })
    listSchemas.mockResolvedValue({ items: [{ name: 'public' }] })

    renderWithProviders(
      <ConnectionSchemaTree item={baseItem('postgres')} onSelectTable={() => {}} />,
    )

    await waitFor(() => expect(screen.getByText('appdb')).toBeInTheDocument())
    fireEvent.click(screen.getByText('appdb'))

    await waitFor(() => expect(listSchemas).toHaveBeenCalledWith('conn-1', 'appdb'))
    expect(listTables).not.toHaveBeenCalled()

    await waitFor(() => expect(screen.getByText('public')).toBeInTheDocument())
    fireEvent.click(screen.getByText('public'))

    await waitFor(() =>
      expect(listTables).toHaveBeenCalledWith('conn-1', { database: 'appdb', schema: 'public' }),
    )
  })
})
