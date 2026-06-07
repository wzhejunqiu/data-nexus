import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import '@/i18n'
import { DataGrid } from './DataGrid'

vi.mock('@/lib/api/table', () => ({
  tableApi: {
    browseRows: vi.fn(),
    updateCellsBatch: vi.fn(),
  },
}))

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    list: vi.fn(),
  },
}))

vi.mock('@/lib/api/schema', () => ({
  schemaApi: {
    getTableSchema: vi.fn(),
  },
}))

vi.mock('@/stores/statusStore', () => ({
  useStatusStore: (selector: (s: { setStatus: () => void }) => unknown) =>
    selector({ setStatus: vi.fn() }),
}))

const openConn = {
  id: 'c1',
  name: 'app.db',
  type: 'sqlite' as const,
  config: { type: 'sqlite' as const, sqlite: { filePath: '/tmp/app.db', readOnly: false } },
  status: 'open' as const,
  lastUsedAt: '2026-01-01T00:00:00Z',
}

const readOnlyConn = {
  ...openConn,
  config: { type: 'sqlite' as const, sqlite: { filePath: '/tmp/app.db', readOnly: true } },
}

const tableData = {
  columns: [
    { name: 'id', dataType: 'INTEGER' },
    { name: 'name', dataType: 'TEXT' },
  ],
  rows: [{ id: 1, name: 'alice' }],
  pagination: { page: 1, pageSize: 50, totalRows: 1, totalPages: 1 },
}

async function setupMocks(readOnly = false) {
  const { connectionApi } = await import('@/lib/api/connection')
  const { schemaApi } = await import('@/lib/api/schema')
  const { tableApi } = await import('@/lib/api/table')
  vi.mocked(connectionApi.list).mockResolvedValue({
    items: [readOnly ? readOnlyConn : openConn],
  })
  vi.mocked(schemaApi.getTableSchema).mockResolvedValue({
    name: 'items',
    type: 'table',
    columns: [
      { name: 'id', dataType: 'INTEGER', primaryKey: true, nullable: false, position: 0 },
      { name: 'name', dataType: 'TEXT', primaryKey: false, nullable: true, position: 1 },
    ],
    indexes: [],
  })
  vi.mocked(tableApi.browseRows).mockResolvedValue(tableData)
}

function renderGrid() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <DataGrid connectionId="c1" tableName="items" />
    </QueryClientProvider>,
  )
}

describe('DataGrid', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('renders NULL and BLOB cells', async () => {
    await setupMocks()
    const { tableApi } = await import('@/lib/api/table')
    vi.mocked(tableApi.browseRows).mockResolvedValue({
      columns: [
        { name: 'name', dataType: 'TEXT' },
        { name: 'data', dataType: 'BLOB' },
      ],
      rows: [
        { name: null, data: { type: 'blob', size: 5 } },
        { name: 'ok', data: null },
      ],
      pagination: { page: 1, pageSize: 50, totalRows: 2, totalPages: 1 },
    })

    renderGrid()

    await waitFor(() => {
      expect(screen.getAllByText('NULL').length).toBeGreaterThan(0)
    })
    const nullCells = screen.getAllByText('NULL')
    nullCells.forEach((cell) => {
      expect(cell.className).toContain('bg-muted/50')
    })
    expect(screen.getByText('[BLOB 5 bytes]')).toBeTruthy()
    expect(screen.getByText('ok')).toBeTruthy()
  })

  it('shows loading state', async () => {
    await setupMocks()
    const { tableApi } = await import('@/lib/api/table')
    vi.mocked(tableApi.browseRows).mockImplementation(() => new Promise(() => {}))

    renderGrid()

    expect(screen.getByTestId('table-skeleton')).toBeTruthy()
  })

  it('shows error state', async () => {
    await setupMocks()
    const { tableApi } = await import('@/lib/api/table')
    vi.mocked(tableApi.browseRows).mockRejectedValue({
      code: 'SQL_ERROR',
      message: 'no such table',
    })

    renderGrid()

    await waitFor(() => {
      expect(screen.getByText('no such table')).toBeTruthy()
    })
  })

  it('hides import on read-only connection', async () => {
    await setupMocks(true)
    renderGrid()
    await waitFor(() => expect(screen.getByText('alice')).toBeTruthy())
    expect(screen.queryByRole('button', { name: /导入 CSV|Import CSV/i })).toBeNull()
  })
})
