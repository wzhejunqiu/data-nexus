import type { DriverType } from '@/lib/types'
import type { ConnectionFormState } from './connectionFormDefaults'

export type ConnectionFormField =
  | 'displayName'
  | 'filePath'
  | 'host'
  | 'port'
  | 'database'
  | 'user'
  | 'password'

export type ConnectionFormErrorCode = 'required' | 'invalidPort' | 'auth' | 'network' | 'database'

export type ConnectionFormErrors = Partial<Record<ConnectionFormField, ConnectionFormErrorCode>>

export type ConnectionFormTranslateFn = (key: string, opts?: Record<string, string>) => string

export function validateConnectionForm(
  state: ConnectionFormState,
  mode: 'create' | 'edit',
  dialect: DriverType,
): ConnectionFormErrors {
  const errors: ConnectionFormErrors = {}

  if (dialect === 'sqlite') {
    // Create mode: empty filePath is allowed; NewConnectionDialog opens the file picker on submit.
    if (mode === 'edit' && !state.filePath.trim()) {
      errors.filePath = 'required'
    }
    return errors
  }

  if (!state.displayName.trim()) {
    errors.displayName = 'required'
  }

  const cfg = dialect === 'postgres' ? state.postgres : state.mysql

  if (!cfg.host.trim()) {
    errors.host = 'required'
  }

  const port = cfg.port
  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    errors.port = 'invalidPort'
  }

  if (dialect === 'postgres' && !cfg.database.trim()) {
    errors.database = 'required'
  }

  if (!cfg.user.trim()) {
    errors.user = 'required'
  }

  if (mode === 'create' && !state.password) {
    errors.password = 'required'
  }

  return errors
}

export function validateConnectionField(
  field: ConnectionFormField,
  state: ConnectionFormState,
  mode: 'create' | 'edit',
  dialect: DriverType,
): ConnectionFormErrorCode | undefined {
  if (dialect === 'sqlite') {
    if (field === 'filePath' && mode === 'edit' && !state.filePath.trim()) {
      return 'required'
    }
    return undefined
  }

  const cfg = dialect === 'postgres' ? state.postgres : state.mysql

  switch (field) {
    case 'displayName':
      return !state.displayName.trim() ? 'required' : undefined
    case 'host':
      return !cfg.host.trim() ? 'required' : undefined
    case 'port': {
      const port = cfg.port
      return !Number.isInteger(port) || port < 1 || port > 65535 ? 'invalidPort' : undefined
    }
    case 'database':
      if (dialect === 'postgres' && !cfg.database.trim()) return 'required'
      return undefined
    case 'user':
      return !cfg.user.trim() ? 'required' : undefined
    case 'password':
      if (mode === 'create' && !state.password) return 'required'
      return undefined
    default:
      return undefined
  }
}

export function connectionFailureFieldErrors(reason: string): ConnectionFormErrors {
  switch (reason) {
    case 'auth':
      return { user: 'auth', password: 'auth' }
    case 'network':
      return { host: 'network', port: 'network' }
    case 'database':
      return { database: 'database' }
    default:
      return {}
  }
}

export function parseConnectionFailureReason(err: unknown): string | null {
  if (
    err &&
    typeof err === 'object' &&
    'code' in err &&
    (err as { code: string }).code === 'CONNECTION_FAILED'
  ) {
    const details = (err as { details?: { reason?: string } }).details
    return details?.reason ?? null
  }
  return null
}

export function formatFieldError(
  t: ConnectionFormTranslateFn,
  field: ConnectionFormField,
  code: ConnectionFormErrorCode | undefined,
): string | undefined {
  if (!code) return undefined
  switch (code) {
    case 'required':
      return t('connectionForm.validation.required', {
        field: t(`connectionForm.fields.${field}`),
      })
    case 'invalidPort':
      return t('connectionForm.validation.invalidPort')
    case 'auth':
      return t('connection.error.auth')
    case 'network':
      return t('connection.error.network')
    case 'database':
      return t('connection.error.database')
    default:
      return undefined
  }
}

export function firstErrorField(errors: ConnectionFormErrors): ConnectionFormField | null {
  const order: ConnectionFormField[] = [
    'displayName',
    'filePath',
    'host',
    'port',
    'database',
    'user',
    'password',
  ]
  for (const field of order) {
    if (errors[field]) return field
  }
  return null
}

export function sectionForField(
  _field: ConnectionFormField,
  _dialect: DriverType,
): 'general' | 'security' | 'advanced' {
  return 'general'
}
