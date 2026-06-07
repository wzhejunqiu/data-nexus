import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import '@/i18n'
import { EditConnectionDialog } from './EditConnectionDialog'
import type { ConnectionListItem } from '@/lib/types'

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    updateSQLiteSettings: vi.fn(),
  },
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
    const checkboxes = screen.getAllByRole('checkbox') as HTMLInputElement[]
    expect(checkboxes[0].checked).toBe(true)
    expect(checkboxes[1].disabled).toBe(true)
  })

  it('calls updateSQLiteSettings when saving toggled readOnly', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.updateSQLiteSettings).mockResolvedValue({} as never)

    renderDialog()
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
})
