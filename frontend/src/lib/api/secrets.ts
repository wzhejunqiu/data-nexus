import {
  ChangeVaultPassword,
  GetSecretsBackend,
  InitVault,
  IsVaultRequiredForRemote,
  LockVault,
  UnlockVault,
  VaultInitialized,
  VaultUnlocked,
} from '../../../wailsjs/go/wails/SecretsService'
import { mapWailsError } from './errors'

async function wrap<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    throw mapWailsError(err)
  }
}

export const secretsApi = {
  getBackend: () => wrap(() => GetSecretsBackend()),
  vaultInitialized: () => wrap(() => VaultInitialized()),
  vaultUnlocked: () => wrap(() => VaultUnlocked()),
  isVaultRequiredForRemote: () => wrap(() => IsVaultRequiredForRemote()),
  initVault: (masterPassword: string) => wrap(() => InitVault(masterPassword)),
  unlockVault: (masterPassword: string) => wrap(() => UnlockVault(masterPassword)),
  lockVault: () => wrap(() => LockVault()),
  changeVaultPassword: (oldPassword: string, newPassword: string) =>
    wrap(() => ChangeVaultPassword(oldPassword, newPassword)),
}

export function isVaultLockedError(err: unknown): boolean {
  const appErr = mapWailsError(err)
  return appErr.code === 'SECRETS_VAULT_LOCKED' || appErr.code === 'SECRETS_VAULT_NOT_INITIALIZED'
}

export function isVaultWrongPasswordError(err: unknown): boolean {
  return mapWailsError(err).code === 'SECRETS_VAULT_WRONG_PASSWORD'
}
