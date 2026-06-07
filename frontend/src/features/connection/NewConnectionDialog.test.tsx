import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { NewConnectionDialog } from './NewConnectionDialog'

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    openFromFile: vi.fn(),
  },
}))

vi.mock('@/lib/api/dialog', () => ({
  dialogApi: {
    openDatabaseFile: vi.fn(),
  },
}))

vi.mock('@tanstack/react-query', () => ({
  useMutation: ({ mutationFn }: { mutationFn: () => Promise<unknown> }) => ({
    mutate: () => mutationFn(),
    isPending: false,
  }),
  useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

vi.mock('@/components/ui/Toast', () => ({
  useToastStore: () => vi.fn(),
}))

describe('NewConnectionDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('passes readOnly flag when opening from path', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.openFromFile).mockResolvedValue({
      id: '1',
      displayName: 'test.db',
    } as never)

    render(<NewConnectionDialog open onOpenChange={() => {}} readOnly />)

    const input = screen.getByPlaceholderText('/path/to/database.db')
    fireEvent.change(input, { target: { value: '/tmp/test.db' } })
    fireEvent.click(screen.getByText('connection.open'))

    expect(connectionApi.openFromFile).toHaveBeenCalledWith({
      filePath: '/tmp/test.db',
      readOnly: true,
    })
  })
})
