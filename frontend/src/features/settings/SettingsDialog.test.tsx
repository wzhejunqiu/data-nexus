import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { SettingsDialogContainer } from './SettingsDialog'

const updateConfig = vi.fn()
const pushToast = vi.fn()

vi.mock('@/lib/api/config', () => ({
  configApi: {
    getConfig: vi.fn().mockResolvedValue({
      log: {
        level: 'info',
        output: 'auto',
        file: {
          path: '',
          max_size_mb: 10,
          max_backups: 3,
          max_age_days: 7,
          compress: false,
        },
      },
    }),
    getConfigPath: vi.fn().mockResolvedValue('/home/user/.data-nexus/config.yaml'),
    updateConfig: (...args: unknown[]) => updateConfig(...args),
  },
}))

vi.mock('@/components/ui/Toast', () => ({
  useToastStore: (selector: (s: { push: typeof pushToast }) => unknown) =>
    selector({ push: pushToast }),
}))

describe('SettingsDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    pushToast.mockClear()
    updateConfig.mockResolvedValue(undefined)
  })

  afterEach(() => {
    cleanup()
  })

  it('shows inline validation for invalid max_size_mb', async () => {
    renderWithProviders(<SettingsDialogContainer open onOpenChange={() => {}} />)

    await waitFor(() => {
      expect(screen.getByDisplayValue('10')).toBeTruthy()
    })

    const sizeInput = screen.getByDisplayValue('10')
    fireEvent.change(sizeInput, { target: { value: '0' } })
    fireEvent.click(screen.getByRole('button', { name: /保存|settings.save/i }))

    await waitFor(() => {
      expect(screen.getByText(/settings.validationInvalid|无效/i)).toBeTruthy()
    })
    expect(updateConfig).not.toHaveBeenCalled()
  })

  it('saves valid settings', async () => {
    renderWithProviders(<SettingsDialogContainer open onOpenChange={() => {}} />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /settings.save|保存/i })).toBeTruthy()
    })

    fireEvent.click(screen.getByRole('button', { name: /settings.save|保存/i }))

    await waitFor(() => {
      expect(updateConfig).toHaveBeenCalledTimes(1)
    })
  })
})
