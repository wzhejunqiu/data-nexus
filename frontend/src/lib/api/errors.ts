import type { TFunction } from 'i18next'
import type { AppError } from '../types'

export type ErrorCode =
  | 'INVALID_REQUEST'
  | 'INVALID_PATH'
  | 'CONNECTION_NOT_FOUND'
  | 'CONNECTION_ALREADY_OPEN'
  | 'CONNECTION_FAILED'
  | 'DATABASE_LOCKED'
  | 'TABLE_NOT_FOUND'
  | 'SQL_ERROR'
  | 'RESULT_TOO_LARGE'
  | 'READ_ONLY'
  | 'DIALOG_CANCELLED'
  | 'EXPORT_CANCELLED'
  | 'EXPORT_NO_STABLE_KEY'
  | 'SAVED_NOT_FOUND'
  | 'SECRETS_VAULT_LOCKED'
  | 'SECRETS_VAULT_NOT_INITIALIZED'
  | 'SECRETS_VAULT_WRONG_PASSWORD'
  | 'INTERNAL_ERROR'

const KNOWN_ERROR_CODES = new Set<string>([
  'INVALID_REQUEST',
  'INVALID_PATH',
  'CONNECTION_NOT_FOUND',
  'CONNECTION_ALREADY_OPEN',
  'CONNECTION_FAILED',
  'DATABASE_LOCKED',
  'TABLE_NOT_FOUND',
  'SQL_ERROR',
  'RESULT_TOO_LARGE',
  'READ_ONLY',
  'DIALOG_CANCELLED',
  'EXPORT_CANCELLED',
  'EXPORT_NO_STABLE_KEY',
  'SAVED_NOT_FOUND',
  'SECRETS_VAULT_LOCKED',
  'SECRETS_VAULT_NOT_INITIALIZED',
  'SECRETS_VAULT_WRONG_PASSWORD',
  'INTERNAL_ERROR',
])

function parseAppErrorString(raw: string): AppError | null {
  const idx = raw.indexOf(': ')
  if (idx <= 0) return null
  const code = raw.slice(0, idx)
  if (!KNOWN_ERROR_CODES.has(code)) return null
  return { code, message: raw.slice(idx + 2) }
}

export function mapWailsError(err: unknown): AppError {
  if (typeof err === 'string') {
    return parseAppErrorString(err) ?? { code: 'INTERNAL_ERROR', message: err }
  }
  if (err && typeof err === 'object') {
    const e = err as Record<string, unknown>
    if (typeof e.code === 'string' && typeof e.message === 'string') {
      return {
        code: e.code,
        message: e.message,
        details: e.details as Record<string, unknown> | undefined,
      }
    }
    if (typeof e.message === 'string') {
      return parseAppErrorString(e.message) ?? { code: 'INTERNAL_ERROR', message: e.message }
    }
  }
  return { code: 'INTERNAL_ERROR', message: String(err) }
}

export function isExportCancelled(err: AppError): boolean {
  return err.code === 'EXPORT_CANCELLED'
}

export function isDialogCancelled(err: AppError): boolean {
  return err.code === 'DIALOG_CANCELLED'
}

export function translateAppError(t: TFunction, err: AppError): string {
  if (err.code === 'SQL_ERROR' && err.message) {
    return err.message
  }
  const key = `errors.${err.code}`
  const translated = t(key, { defaultValue: '' })
  if (translated) return translated
  return err.message || t('errors.INTERNAL_ERROR')
}

export function formatError(t: TFunction, err: unknown): string {
  const appErr = mapWailsError(err)
  return translateAppError(t, appErr)
}
