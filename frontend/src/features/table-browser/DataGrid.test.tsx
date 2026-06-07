import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import '@/i18n'
import { DataGrid } from './DataGrid'

vi.mock('@/lib/api/table', () => ({
  tableApi: {
    browseRows: vi.fn(),
  },
}))

vi.mock('@/stores/statusStore', () => ({
  useStatusStore: (selector: (s: { setStatus: () => void }) => unknown) =>
    selector({ setStatus: vi.fn() }),
}))

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
    expect(screen.getByText('[BLOB 5 bytes]')).toBeTruthy()
    expect(screen.getByText('ok')).toBeTruthy()
  })

  it('shows loading state', async () => {
    const { tableApi } = await import('@/lib/api/table')
    vi.mocked(tableApi.browseRows).mockImplementation(() => new Promise(() => {}))

    renderGrid()

    expect(screen.getByText(/加载/)).toBeTruthy()
  })

  it('shows error state', async () => {
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
})
