import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ColumnProfile } from './ColumnProfile'

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
          topValues: [{ value: 'a', count: 2 }],
        },
      ],
    }),
  },
}))

describe('ColumnProfile', () => {
  it('renders profile stats', async () => {
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={qc}>
        <ColumnProfile connectionId="c1" tableName="items" />
      </QueryClientProvider>,
    )
    const row = await screen.findByRole('row', { name: /category/i })
    expect(row).toBeInTheDocument()
    expect(row.textContent).toContain('2')
  })
})
