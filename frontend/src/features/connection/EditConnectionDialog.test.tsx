import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import '@/i18n'
import { EditConnectionDialog } from './EditConnectionDialog'
import type { ConnectionListItem } from '@/lib/types'

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    updateReadOnly: vi.fn(),
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
    sqlite: { filePath: '/tmp/test.db', readOnly: false },
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

  it('renders readOnly checkbox with initial value', () => {
    renderDialog()
    const checkbox = screen.getByRole('checkbox') as HTMLInputElement
    expect(checkbox.checked).toBe(false)
  })

  it('shows readOnly checked when connection is read-only', () => {
    renderDialog({
      ...baseItem,
      config: {
        type: 'sqlite',
        sqlite: { filePath: '/tmp/test.db', readOnly: true },
      },
    })
    expect((screen.getByRole('checkbox') as HTMLInputElement).checked).toBe(true)
  })

  it('calls updateReadOnly when saving toggled checkbox', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.updateReadOnly).mockResolvedValue({} as never)

    renderDialog()
    fireEvent.click(screen.getByRole('checkbox'))
    fireEvent.click(screen.getByRole('button', { name: /confirm|确认/i }))

    await waitFor(() => {
      expect(connectionApi.updateReadOnly).toHaveBeenCalledWith('conn-1', true)
    })
  })
})
