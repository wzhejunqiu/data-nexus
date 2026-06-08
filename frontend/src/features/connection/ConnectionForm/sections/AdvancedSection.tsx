import { useTranslation } from 'react-i18next'
import type { DriverType } from '@/lib/types'
import { PostgresAdvancedFields } from '../fields/PostgresFields'
import { MySQLAdvancedFields } from '../fields/MySQLFields'
import type { ConnectionFormState } from '../connectionFormDefaults'

export function AdvancedSection({
  dialect,
  state,
  onChange,
}: {
  dialect: DriverType
  state: ConnectionFormState
  onChange: (patch: Partial<ConnectionFormState>) => void
}) {
  const { t } = useTranslation()

  if (dialect === 'sqlite') {
    return (
      <div className="space-y-3">
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={state.readOnly}
            onChange={(e) => {
              const next = e.target.checked
              onChange({ readOnly: next, wal: next ? false : state.wal })
            }}
          />
          {t('connection.readOnly')}
        </label>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={state.wal}
            disabled={state.readOnly}
            onChange={(e) => onChange({ wal: e.target.checked })}
          />
          {t('connection.wal')}
        </label>
      </div>
    )
  }
  if (dialect === 'postgres') {
    return <PostgresAdvancedFields state={state} onChange={onChange} />
  }
  if (dialect === 'mysql') {
    return <MySQLAdvancedFields state={state} onChange={onChange} />
  }
  return null
}
