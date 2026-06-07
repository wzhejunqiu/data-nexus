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
  | 'SAVED_NOT_FOUND'
  | 'INTERNAL_ERROR'

export function mapWailsError(err: unknown): AppError {
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
      return { code: 'INTERNAL_ERROR', message: e.message }
    }
  }
  return { code: 'INTERNAL_ERROR', message: String(err) }
}

export function isDialogCancelled(err: AppError): boolean {
  return err.code === 'DIALOG_CANCELLED'
}

export function translateAppError(t: TFunction, err: AppError): string {
  const key = `errors.${err.code}`
  const translated = t(key, { defaultValue: '' })
  if (translated) return translated
  return err.message || t('errors.INTERNAL_ERROR')
}

export function formatError(t: TFunction, err: unknown): string {
  const appErr = mapWailsError(err)
  return translateAppError(t, appErr)
}
