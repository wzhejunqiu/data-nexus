import { describe, expect, it, vi, beforeEach } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { ConnectionSchemaTree } from './ConnectionSchemaTree'
import type { ConnectionListItem } from '@/lib/types'
import { useWorkspaceStore } from '@/stores/workspaceStore'

const listNamespaces = vi.fn()
const listSchemas = vi.fn()
const listTables = vi.fn()
const detach = vi.fn()

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
    detach: (...args: unknown[]) => detach(...args),
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

function expandNamespace(name: string) {
  const row = screen.getByText(name).closest('div')!
  const chevron = row.querySelector('button[aria-label="expand"], button[aria-label="collapse"]')!
  fireEvent.click(chevron)
}

describe('ConnectionSchemaTree lazy loading', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useWorkspaceStore.setState({
      browseContext: {},
      activeConnectionId: null,
      selectedTable: null,
    })
    listNamespaces.mockResolvedValue({ items: [] })
    listSchemas.mockResolvedValue({ items: [] })
    listTables.mockResolvedValue({ items: [{ name: 'users', type: 'table' }] })
    detach.mockResolvedValue(undefined)
  })

  it('loads namespaces on mount only for sqlite', async () => {
    listNamespaces.mockResolvedValue({ items: [{ name: 'main', kind: 'main' }] })

    renderWithProviders(<ConnectionSchemaTree item={baseItem('sqlite')} onSelectTable={() => {}} />)

    await waitFor(() => expect(listNamespaces).toHaveBeenCalledWith('conn-1'))
    expect(listTables).not.toHaveBeenCalled()
    expect(listSchemas).not.toHaveBeenCalled()
  })

  it('sets browseContext when clicking namespace row', async () => {
    listNamespaces.mockResolvedValue({ items: [{ name: 'main', kind: 'main' }] })

    renderWithProviders(<ConnectionSchemaTree item={baseItem('sqlite')} onSelectTable={() => {}} />)

    await waitFor(() => expect(screen.getByText('main')).toBeInTheDocument())
    fireEvent.click(screen.getByText('main'))

    expect(useWorkspaceStore.getState().browseContext['conn-1']).toEqual({ database: 'main' })
    expect(listTables).not.toHaveBeenCalled()
  })

  it('loads tables after expanding sqlite namespace via chevron', async () => {
    listNamespaces.mockResolvedValue({ items: [{ name: 'main', kind: 'main' }] })

    renderWithProviders(<ConnectionSchemaTree item={baseItem('sqlite')} onSelectTable={() => {}} />)

    await waitFor(() => expect(screen.getByText('main')).toBeInTheDocument())
    expandNamespace('main')

    await waitFor(() =>
      expect(listTables).toHaveBeenCalledWith('conn-1', { database: 'main', schema: undefined }),
    )
    expect(listSchemas).not.toHaveBeenCalled()
  })

  it('loads tables after expanding mysql namespace', async () => {
    listNamespaces.mockResolvedValue({ items: [{ name: 'app', kind: 'database' }] })

    renderWithProviders(<ConnectionSchemaTree item={baseItem('mysql')} onSelectTable={() => {}} />)

    await waitFor(() => expect(screen.getByText('app')).toBeInTheDocument())
    expandNamespace('app')

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
    expandNamespace('app')
    await waitFor(() => expect(screen.getByRole('button', { name: /users/ })).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: /users/ }))

    expect(onSelectTable).toHaveBeenCalledWith('app.users', { database: 'app', schema: undefined })
  })

  it('renders TABLES group header', async () => {
    listNamespaces.mockResolvedValue({ items: [{ name: 'main', kind: 'main' }] })
    listTables.mockResolvedValue({
      items: [
        { name: 'users', type: 'table' },
        { name: 'active', type: 'view' },
      ],
    })

    renderWithProviders(<ConnectionSchemaTree item={baseItem('sqlite')} onSelectTable={() => {}} />)

    await waitFor(() => expect(screen.getByText('main')).toBeInTheDocument())
    expandNamespace('main')

    await waitFor(() => expect(screen.getByText(/TABLES \(1\)|表 \(1\)/)).toBeInTheDocument())
    await waitFor(() => expect(screen.getByText(/VIEWS \(1\)|视图 \(1\)/)).toBeInTheDocument())
  })

  it('weakens system database namespace style', async () => {
    listNamespaces.mockResolvedValue({
      items: [{ name: 'information_schema', kind: 'database' }],
    })

    renderWithProviders(<ConnectionSchemaTree item={baseItem('mysql')} onSelectTable={() => {}} />)

    await waitFor(() => expect(screen.getByText('information_schema')).toBeInTheDocument())
    const row = screen.getByText('information_schema').closest('div')!
    expect(row.className).toMatch(/italic/)
    expect(row.className).toMatch(/text-muted/)
  })

  it('loads schemas then tables for postgres', async () => {
    listNamespaces.mockResolvedValue({ items: [{ name: 'appdb', kind: 'database' }] })
    listSchemas.mockResolvedValue({ items: [{ name: 'public' }] })

    renderWithProviders(
      <ConnectionSchemaTree item={baseItem('postgres')} onSelectTable={() => {}} />,
    )

    await waitFor(() => expect(screen.getByText('appdb')).toBeInTheDocument())
    expandNamespace('appdb')

    await waitFor(() => expect(listSchemas).toHaveBeenCalledWith('conn-1', 'appdb'))
    expect(listTables).not.toHaveBeenCalled()

    await waitFor(() => expect(screen.getByText('public')).toBeInTheDocument())
    fireEvent.click(screen.getByText('public'))
    expect(useWorkspaceStore.getState().browseContext['conn-1']).toEqual({
      database: 'appdb',
      schema: 'public',
    })

    expandNamespace('public')
    await waitFor(() =>
      expect(listTables).toHaveBeenCalledWith('conn-1', { database: 'appdb', schema: 'public' }),
    )
  })

  it('confirms detach when browseContext matches attach alias', async () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)
    listNamespaces.mockResolvedValue({
      items: [{ name: 'logs', kind: 'attach', filePath: '/tmp/logs.db' }],
    })
    useWorkspaceStore.setState({
      browseContext: { 'conn-1': { database: 'logs' } },
    })

    renderWithProviders(<ConnectionSchemaTree item={baseItem('sqlite')} onSelectTable={() => {}} />)

    await waitFor(() => expect(screen.getByText('logs')).toBeInTheDocument())
    fireEvent.contextMenu(screen.getByText('logs'))
    const detachItem = await screen.findByText(/取消附加|Detach/i)
    fireEvent.click(detachItem)

    expect(confirmSpy).toHaveBeenCalled()
    expect(detach).not.toHaveBeenCalled()
    confirmSpy.mockRestore()
  })

  it('detaches without confirm when browseContext does not match', async () => {
    const confirmSpy = vi.spyOn(window, 'confirm')
    listNamespaces.mockResolvedValue({
      items: [{ name: 'logs', kind: 'attach', filePath: '/tmp/logs.db' }],
    })

    renderWithProviders(<ConnectionSchemaTree item={baseItem('sqlite')} onSelectTable={() => {}} />)

    await waitFor(() => expect(screen.getByText('logs')).toBeInTheDocument())
    fireEvent.contextMenu(screen.getByText('logs'))
    const detachItem = await screen.findByText(/取消附加|Detach/i)
    fireEvent.click(detachItem)

    expect(confirmSpy).not.toHaveBeenCalled()
    await waitFor(() => expect(detach).toHaveBeenCalledWith('conn-1', 'logs'))
    confirmSpy.mockRestore()
  })
})
