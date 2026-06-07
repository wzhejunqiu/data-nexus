import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { renderWithProviders } from '@/test/render'
import { SavedQueries } from './SavedQueries'
import { useWorkspaceStore } from '@/stores/workspaceStore'

vi.mock('@/lib/api/queries', () => ({
  queriesApi: {
    list: vi.fn().mockResolvedValue({ items: [] }),
    save: vi.fn().mockResolvedValue({ id: 'q1', name: 'Test', sql: 'SELECT 42;' }),
    delete: vi.fn(),
  },
}))

vi.mock('@/components/ui/Toast', () => ({
  useToastStore: (sel: (s: { push: () => void }) => unknown) => sel({ push: vi.fn() }),
}))

describe('SavedQueries', () => {
  beforeEach(() => {
    useWorkspaceStore.setState({
      editorSql: 'SELECT 42;',
      pendingSql: null,
      activeConnectionId: 'c1',
    })
  })

  it('saves editorSql via save dialog', async () => {
    const { queriesApi } = await import('@/lib/api/queries')
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    renderWithProviders(
      <QueryClientProvider client={qc}>
        <SavedQueries />
      </QueryClientProvider>,
    )

    fireEvent.click(await screen.findByRole('button', { name: /保存|save/i }))
    const input = await screen.findByRole('textbox')
    fireEvent.change(input, { target: { value: 'My saved query' } })
    fireEvent.click(screen.getByRole('button', { name: /确认|confirm/i }))

    await waitFor(() => {
      expect(queriesApi.save).toHaveBeenCalledWith({
        name: 'My saved query',
        sql: 'SELECT 42;',
        connectionId: 'c1',
      })
    })
  })
})
