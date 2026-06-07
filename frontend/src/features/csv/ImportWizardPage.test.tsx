import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { ImportWizardPage } from './ImportWizardPage'
import { useImportStore } from '@/stores/importStore'

vi.mock('@/lib/api/dialog', () => ({
  dialogApi: {
    openCSVFile: vi.fn().mockResolvedValue('/tmp/data.csv'),
  },
}))

vi.mock('@/lib/api/import', () => ({
  importApi: {
    parseCSVPreview: vi.fn(),
    importCSV: vi.fn(),
  },
}))

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    list: vi.fn(),
  },
}))

vi.mock('@/features/sql-editor/SqlCodeView', () => ({
  SqlCodeView: ({ sql }: { sql: string }) => <div data-testid="sql-code-view">{sql}</div>,
}))

vi.mock('@/lib/api/schema', () => ({
  schemaApi: {
    listTables: vi.fn(),
    getTableSchema: vi.fn(),
  },
}))

const pushToast = vi.fn()
vi.mock('@/components/ui/Toast', () => ({
  useToastStore: (selector: (s: { push: typeof pushToast }) => unknown) =>
    selector({ push: pushToast }),
}))

describe('ImportWizardPage', () => {
  beforeEach(async () => {
    vi.clearAllMocks()
    useImportStore.getState().closeImport()
    useImportStore.getState().openImport({ connectionId: 'c1', defaultTable: 'items' })

    const { importApi } = await import('@/lib/api/import')
    const { connectionApi } = await import('@/lib/api/connection')
    const { schemaApi } = await import('@/lib/api/schema')
    vi.mocked(connectionApi.list).mockResolvedValue({
      items: [
        {
          id: 'c1',
          name: 'app.db',
          type: 'sqlite',
          config: { type: 'sqlite', sqlite: { filePath: '/tmp/app.db', readOnly: false } },
          status: 'open',
          lastUsedAt: '2026-01-01T00:00:00Z',
        },
      ],
    })
    vi.mocked(importApi.parseCSVPreview).mockResolvedValue({
      headers: ['id', 'name'],
      rows: [{ id: '1', name: 'alice' }],
      rowCount: 1,
      hasHeader: true,
    })
    vi.mocked(schemaApi.listTables).mockResolvedValue({
      items: [{ name: 'items', type: 'table' }],
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
  })

  afterEach(() => {
    cleanup()
    useImportStore.getState().closeImport()
  })

  it('auto-maps CSV headers to existing table columns in mapping step', async () => {
    renderWithProviders(<ImportWizardPage />)

    fireEvent.click(screen.getByRole('button', { name: /选择文件|Choose file/i }))
    await waitFor(() => {
      expect(screen.getByText('/tmp/data.csv')).toBeTruthy()
    })

    fireEvent.click(screen.getByRole('button', { name: /预览|Preview/i }))
    await waitFor(() => {
      expect(screen.getByText('alice')).toBeTruthy()
    })

    fireEvent.click(screen.getByRole('button', { name: /下一步|Next/i }))

    await waitFor(() => {
      const selects = screen.getAllByRole('combobox')
      expect(selects).toHaveLength(2)
      expect((selects[0] as HTMLSelectElement).value).toBe('id')
      expect((selects[1] as HTMLSelectElement).value).toBe('name')
    })
  })

  it('shows step indicator', () => {
    renderWithProviders(<ImportWizardPage />)
    expect(screen.getByText(/第 1 \/ 3 步|Step 1 of 3/i)).toBeTruthy()
  })

  it('shows type and primary key controls for new table mapping', async () => {
    renderWithProviders(<ImportWizardPage />)

    fireEvent.click(screen.getByRole('button', { name: /选择文件|Choose file/i }))
    await waitFor(() => expect(screen.getByText('/tmp/data.csv')).toBeTruthy())

    fireEvent.click(screen.getByRole('button', { name: /预览|Preview/i }))
    await waitFor(() => expect(screen.getByText('alice')).toBeTruthy())

    fireEvent.click(screen.getByLabelText(/新建表|Create new table/i))
    fireEvent.change(screen.getByPlaceholderText(/新表名|New table name/i), {
      target: { value: 'people' },
    })
    fireEvent.click(screen.getByRole('button', { name: /下一步|Next/i }))

    await waitFor(() => {
      expect(screen.getByRole('columnheader', { name: /列名|Column name/i })).toBeTruthy()
      expect(screen.getByRole('columnheader', { name: /主键|Primary key/i })).toBeTruthy()
    })

    const pkCheckbox = screen.getAllByRole('checkbox')[0] as HTMLInputElement
    fireEvent.click(pkCheckbox)
    expect(pkCheckbox.checked).toBe(true)
  })

  it('shows import duration on success result', async () => {
    const { importApi } = await import('@/lib/api/import')
    vi.mocked(importApi.importCSV).mockResolvedValue({ rowsInserted: 2, rowsUpdated: 0 })

    renderWithProviders(<ImportWizardPage />)

    fireEvent.click(screen.getByRole('button', { name: /选择文件|Choose file/i }))
    await waitFor(() => expect(screen.getByText('/tmp/data.csv')).toBeTruthy())

    fireEvent.click(screen.getByRole('button', { name: /预览|Preview/i }))
    await waitFor(() => expect(screen.getByText('alice')).toBeTruthy())

    fireEvent.click(screen.getByRole('button', { name: /下一步|Next/i }))
    await waitFor(() => expect(screen.getAllByRole('combobox')).toHaveLength(2))

    fireEvent.click(screen.getByRole('button', { name: /开始导入|Start import/i }))
    await waitFor(() => {
      expect(screen.getByText(/导入成功|Import completed/i)).toBeTruthy()
      expect(screen.getByText(/耗时|Duration/i)).toBeTruthy()
    })
  })

  it('returns to mapping step after import error so user can edit', async () => {
    const { importApi } = await import('@/lib/api/import')
    vi.mocked(importApi.importCSV).mockRejectedValue(new Error('duplicate key'))

    renderWithProviders(<ImportWizardPage />)

    fireEvent.click(screen.getByRole('button', { name: /选择文件|Choose file/i }))
    await waitFor(() => expect(screen.getByText('/tmp/data.csv')).toBeTruthy())

    fireEvent.click(screen.getByRole('button', { name: /预览|Preview/i }))
    await waitFor(() => expect(screen.getByText('alice')).toBeTruthy())

    fireEvent.click(screen.getByRole('button', { name: /下一步|Next/i }))
    await waitFor(() => expect(screen.getAllByRole('combobox')).toHaveLength(2))

    fireEvent.click(screen.getByRole('button', { name: /开始导入|Start import/i }))
    await waitFor(() => {
      expect(screen.getByText(/导入失败|Import failed/i)).toBeTruthy()
    })

    fireEvent.click(screen.getByRole('button', { name: /上一步|Back/i }))
    await waitFor(() => {
      expect(screen.getByText(/第 3 \/ 3 步|Step 3 of 3/i)).toBeTruthy()
      expect(screen.getByRole('button', { name: /开始导入|Start import/i })).toBeTruthy()
    })
  })

  it('shows CREATE TABLE preview on mapping step for new table', async () => {
    renderWithProviders(<ImportWizardPage />)

    fireEvent.click(screen.getByRole('button', { name: /选择文件|Choose file/i }))
    await waitFor(() => expect(screen.getByText('/tmp/data.csv')).toBeTruthy())

    fireEvent.click(screen.getByRole('button', { name: /预览|Preview/i }))
    await waitFor(() => expect(screen.getByText('alice')).toBeTruthy())

    fireEvent.click(screen.getByLabelText(/新建表|Create new table/i))
    fireEvent.change(screen.getByPlaceholderText(/新表名|New table name/i), {
      target: { value: 'people' },
    })
    fireEvent.click(screen.getByRole('button', { name: /下一步|Next/i }))
    await waitFor(() => {
      expect(screen.getByRole('columnheader', { name: /列名|Column name/i })).toBeTruthy()
      expect(screen.getByTestId('sql-code-view')).toBeTruthy()
      expect(screen.getByTestId('sql-code-view').textContent).toContain('CREATE TABLE "people"')
    })
  })
})
