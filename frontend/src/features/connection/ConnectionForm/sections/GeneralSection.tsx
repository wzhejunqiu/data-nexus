import type { DriverType } from '@/lib/types'
import { SQLiteFields } from '../fields/SQLiteFields'
import { PostgresFields } from '../fields/PostgresFields'
import { MySQLFields } from '../fields/MySQLFields'
import type { ConnectionFormErrors } from '../connectionFormValidation'
import type { ConnectionFormState } from '../connectionFormDefaults'

export function GeneralSection({
  dialect,
  mode,
  state,
  errors,
  onChange,
  onBrowse,
}: {
  dialect: DriverType
  mode: 'create' | 'edit'
  state: ConnectionFormState
  errors: ConnectionFormErrors
  onChange: (patch: Partial<ConnectionFormState>) => void
  onBrowse?: () => Promise<string | undefined>
}) {
  if (dialect === 'sqlite') {
    return (
      <SQLiteFields
        state={state}
        errors={errors}
        mode={mode}
        onChange={onChange}
        onBrowse={onBrowse}
      />
    )
  }
  if (dialect === 'postgres') {
    return <PostgresFields state={state} errors={errors} mode={mode} onChange={onChange} />
  }
  return <MySQLFields state={state} errors={errors} mode={mode} onChange={onChange} />
}
