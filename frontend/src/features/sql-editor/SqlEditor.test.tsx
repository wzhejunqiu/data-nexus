import { describe, expect, it, vi, beforeEach } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { SqlEditor } from './SqlEditor'
import { useQueryHistoryStore } from '@/stores/queryHistoryStore'
import { useStatusStore } from '@/stores/statusStore'

vi.mock('@monaco-editor/react', () => ({
  default: ({
    value,
    onChange,
  }: {
    value: string
    onChange?: (v: string | undefined) => void
  }) => (
    <textarea
      data-testid="sql-editor"
      value={value}
      onChange={(e) => onChange?.(e.target.value)}
    />
  ),
}))

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    list: vi.fn(),
  },
}))

vi.mock('@/lib/api/query', () => ({
  queryApi: {
    execute: vi.fn(),
    classifySQL: vi.fn(),
  },
}))

const openConn = {
  id: 'c1',
  name: 'app.db',
  type: 'sqlite' as const,
  config: { type: 'sqlite' as const, sqlite: { filePath: '/tmp/app.db', readOnly: false } },
  status: 'open' as const,
  lastUsedAt: '2026-01-01T00:00:00Z',
}

describe('SqlEditor', () => {
  beforeEach(() => {
    useQueryHistoryStore.setState({ items: {} })
    useStatusStore.setState({ rowCount: null, durationMs: null, operation: null })
    vi.mocked(navigator.clipboard.writeText).mockClear()
  })

  it('disables run without connection', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [] })

    renderWithProviders(<SqlEditor connectionId={null} />)
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /执行/ })).toBeDisabled()
    })
  })

  it('runs SELECT and shows result', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    const { queryApi } = await import('@/lib/api/query')
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [openConn] })
    vi.mocked(queryApi.classifySQL).mockResolvedValue('query')
    vi.mocked(queryApi.execute).mockResolvedValue({
      kind: 'result',
      columns: [{ name: 'n', dataType: 'INTEGER' }],
      rows: [{ n: 1 }],
      rowCount: 1,
      durationMs: 12,
    })

    renderWithProviders(<SqlEditor connectionId="c1" />)
    fireEvent.click(screen.getByRole('button', { name: /执行/ }))

    await waitFor(() => {
      expect(queryApi.execute).toHaveBeenCalled()
      expect(screen.getByText('1')).toBeInTheDocument()
    })
    expect(useQueryHistoryStore.getState().items.c1).toContain('SELECT 1;')
    expect(useStatusStore.getState().operation).toBe('query')
  })

  it('shows confirm dialog for write sql', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    const { queryApi } = await import('@/lib/api/query')
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [openConn] })
    vi.mocked(queryApi.classifySQL).mockResolvedValue('write')

    renderWithProviders(<SqlEditor connectionId="c1" />)
    fireEvent.change(screen.getByTestId('sql-editor'), { target: { value: 'DELETE FROM t' } })
    fireEvent.click(screen.getByRole('button', { name: /执行/ }))

    await waitFor(() => {
      expect(screen.getByText(/确认执行此写操作/)).toBeInTheDocument()
    })
    expect(queryApi.execute).not.toHaveBeenCalled()
  })

  it('blocks write on read-only connection', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    const { queryApi } = await import('@/lib/api/query')
    vi.mocked(connectionApi.list).mockResolvedValue({
      items: [
        {
          ...openConn,
          config: { type: 'sqlite', sqlite: { filePath: '/tmp/app.db', readOnly: true } },
        },
      ],
    })
    vi.mocked(queryApi.classifySQL).mockResolvedValue('write')

    renderWithProviders(<SqlEditor connectionId="c1" />)
    await waitFor(() => {
      expect(screen.getByRole('option', { name: 'app.db' })).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: /执行/ }))

    await waitFor(() => {
      expect(screen.getByText('当前连接为只读模式')).toBeInTheDocument()
    })
    expect(queryApi.execute).not.toHaveBeenCalled()
  })

  it('shows classify error', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    const { queryApi } = await import('@/lib/api/query')
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [openConn] })
    vi.mocked(queryApi.classifySQL).mockRejectedValue({ code: 'SQL_ERROR', message: 'bad sql' })

    renderWithProviders(<SqlEditor connectionId="c1" />)
    fireEvent.click(screen.getByRole('button', { name: /执行/ }))

    await waitFor(() => {
      expect(screen.getByText('bad sql')).toBeInTheDocument()
    })
  })

  it('updates sql from pragma dropdown', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [openConn] })

    renderWithProviders(<SqlEditor connectionId="c1" />)
    const selects = screen.getAllByRole('combobox')
    const pragmaSelect = selects[selects.length - 1]
    fireEvent.change(pragmaSelect, {
      target: { value: 'PRAGMA foreign_keys;' },
    })
    expect((screen.getByTestId('sql-editor') as HTMLTextAreaElement).value).toBe(
      'PRAGMA foreign_keys;',
    )
  })

  it('shows exec result and copies csv', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    const { queryApi } = await import('@/lib/api/query')
    vi.mocked(connectionApi.list).mockResolvedValue({ items: [openConn] })
    vi.mocked(queryApi.classifySQL).mockResolvedValue('query')
    vi.mocked(queryApi.execute).mockResolvedValue({
      kind: 'result',
      columns: [{ name: 'id', dataType: 'INTEGER' }],
      rows: [{ id: 7 }],
      rowCount: 1,
      durationMs: 3,
    })

    renderWithProviders(<SqlEditor connectionId="c1" />)
    fireEvent.click(screen.getByRole('button', { name: /执行/ }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /复制为 CSV/ })).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: /复制为 CSV/ }))
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('id\n7')
  })
})
