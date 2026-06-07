import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { ImportWizard } from './ImportWizard'

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

describe('ImportWizard', () => {
  beforeEach(async () => {
    vi.clearAllMocks()
    const { importApi } = await import('@/lib/api/import')
    const { schemaApi } = await import('@/lib/api/schema')
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
  })

  it('auto-maps CSV headers to existing table columns in mapping step', async () => {
    renderWithProviders(
      <ImportWizard connectionId="c1" defaultTable="items" onClose={() => {}} onDone={() => {}} />,
    )

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
})
