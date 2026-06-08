import type { DriverType } from '@/lib/types'
import { PostgresSecurityFields } from '../fields/PostgresFields'
import { MySQLSecurityFields } from '../fields/MySQLFields'
import type { ConnectionFormState } from '../connectionFormDefaults'

export function SecuritySection({
  dialect,
  state,
  onChange,
}: {
  dialect: DriverType
  state: ConnectionFormState
  onChange: (patch: Partial<ConnectionFormState>) => void
}) {
  if (dialect === 'postgres') {
    return <PostgresSecurityFields state={state} onChange={onChange} />
  }
  if (dialect === 'mysql') {
    return <MySQLSecurityFields state={state} onChange={onChange} />
  }
  return null
}
