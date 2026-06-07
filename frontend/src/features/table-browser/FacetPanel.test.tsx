import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { FacetPanel } from './FacetPanel'

vi.mock('@/lib/api/schema', () => ({
  schemaApi: {
    getTableProfile: vi.fn().mockResolvedValue({
      tableName: 'items',
      sampledRows: 3,
      totalRows: 3,
      isSampled: false,
      columns: [
        {
          name: 'category',
          distinctCount: 2,
          nullPercent: 0,
          isLowCardinality: true,
          topValues: [
            { value: 'a', count: 2 },
            { value: 'b', count: 1 },
          ],
        },
      ],
    }),
  },
}))

describe('FacetPanel', () => {
  it('renders facet values and toggles filter', async () => {
    const onToggleFilter = vi.fn()
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={qc}>
        <FacetPanel
          connectionId="c1"
          tableName="items"
          filters={[]}
          onToggleFilter={onToggleFilter}
        />
      </QueryClientProvider>,
    )

    fireEvent.click(await screen.findByRole('button', { name: /facet/i }))
    fireEvent.click(await screen.findByRole('button', { name: /a \(2\)/i }))
    expect(onToggleFilter).toHaveBeenCalledWith({
      column: 'category',
      operator: 'eq',
      value: 'a',
    })
  })
})
