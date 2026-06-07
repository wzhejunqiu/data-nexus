import type { AppError } from '../types'

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
