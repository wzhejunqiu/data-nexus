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

export type ConnectionFormErrors = Partial<Record<ConnectionFormField, string>>

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
  field: ConnectionFormField,
  dialect: DriverType,
): 'general' | 'security' | 'advanced' {
  if (field === 'password' && dialect !== 'sqlite') return 'general'
  if (
    field === 'displayName' ||
    field === 'filePath' ||
    field === 'host' ||
    field === 'port' ||
    field === 'database' ||
    field === 'user'
  ) {
    return 'general'
  }
  return 'general'
}
