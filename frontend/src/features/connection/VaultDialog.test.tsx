import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, render, screen, fireEvent, waitFor } from '@testing-library/react'
import { VaultDialog } from './VaultDialog'

vi.mock('@/lib/api/secrets', () => ({
  secretsApi: {
    initVault: vi.fn(),
    unlockVault: vi.fn(),
  },
  isVaultWrongPasswordError: (err: unknown) =>
    err &&
    typeof err === 'object' &&
    (err as { code?: string }).code === 'SECRETS_VAULT_WRONG_PASSWORD',
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

vi.mock('@/components/ui/Toast', () => ({
  useToastStore: (selector: (s: { push: ReturnType<typeof vi.fn> }) => unknown) =>
    selector({ push: vi.fn() }),
}))

describe('VaultDialog', () => {
  beforeEach(() => vi.clearAllMocks())
  afterEach(() => cleanup())

  it('submits unlock when password is long enough', async () => {
    const { secretsApi } = await import('@/lib/api/secrets')
    vi.mocked(secretsApi.unlockVault).mockResolvedValue(undefined)
    const onSuccess = vi.fn()

    render(<VaultDialog open mode="unlock" onOpenChange={() => {}} onSuccess={onSuccess} />)

    fireEvent.change(screen.getByPlaceholderText('vault.masterPassword'), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'vault.unlockAction' }))

    await waitFor(() => {
      expect(secretsApi.unlockVault).toHaveBeenCalledWith('password123')
      expect(onSuccess).toHaveBeenCalled()
    })
  })
})
