import type { RefObject } from 'react'
import type { TFunction } from 'i18next'
import { formatError } from '@/lib/api/errors'
import type { ConnectionFormHandle } from './ConnectionForm/ConnectionForm'
import { parseConnectionFailureReason } from './ConnectionForm/connectionFormValidation'

export function formatConnectionErrorMessage(t: TFunction, err: unknown): string {
  const reason = parseConnectionFailureReason(err)
  if (reason === 'auth') return t('connection.error.auth')
  if (reason === 'network') return t('connection.error.network')
  if (reason === 'database') return t('connection.error.database')
  return formatError(t, err)
}

export function handleRemoteConnectionError(
  err: unknown,
  {
    t,
    pushToast,
    formRef,
  }: {
    t: TFunction
    pushToast: (message: string, variant: 'error' | 'success') => void
    formRef: RefObject<ConnectionFormHandle | null>
  },
): boolean {
  const reason = parseConnectionFailureReason(err)
  if (!reason) return false
  pushToast(formatConnectionErrorMessage(t, err), 'error')
  formRef.current?.highlightConnectionFailure(reason)
  return true
}
