import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, screen, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { AboutDialog } from './AboutDialog'

vi.mock('@/lib/api/app', () => ({
  appApi: {
    getVersion: vi.fn(),
    getPlatform: vi.fn(),
  },
}))

describe('AboutDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('shows version and platform when open', async () => {
    const { appApi } = await import('@/lib/api/app')
    vi.mocked(appApi.getVersion).mockResolvedValue({
      version: '0.5.0',
      platform: 'darwin',
    } as never)
    vi.mocked(appApi.getPlatform).mockResolvedValue('darwin')

    renderWithProviders(<AboutDialog open onOpenChange={() => {}} />)

    await waitFor(() => {
      expect(screen.getByText('Data Nexus')).toBeInTheDocument()
      expect(screen.getByText(/0\.5\.0/)).toBeInTheDocument()
      expect(screen.getByText(/darwin/)).toBeInTheDocument()
    })
  })
})
