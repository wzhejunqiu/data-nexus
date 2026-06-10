import { useTranslation } from 'react-i18next'
import { Input } from '@/components/ui/Input'
import { fieldErrorClass, FormField } from '../FormField'
import { MYSQL_COLLATIONS } from '../connectionFormDefaults'
import {
  formatFieldError,
  type ConnectionFormErrors,
  type ConnectionFormField,
} from '../connectionFormValidation'
import type { ConnectionFormState } from '../connectionFormDefaults'

export function MySQLFields({
  state,
  errors,
  mode,
  onChange,
  onFieldBlur,
}: {
  state: ConnectionFormState
  errors: ConnectionFormErrors
  mode: 'create' | 'edit'
  onChange: (patch: Partial<ConnectionFormState>) => void
  onFieldBlur?: (field: ConnectionFormField) => void
}) {
  const { t } = useTranslation()
  const my = state.mysql
  const optionalLabel = t('connectionForm.optional')

  const setMySQL = (patch: Partial<typeof my>) => onChange({ mysql: { ...my, ...patch } })

  return (
    <div className="space-y-3">
      <FormField
        id="displayName"
        label={t('connectionForm.fields.displayName')}
        required
        error={formatFieldError(t, 'displayName', errors.displayName)}
      >
        <Input
          id="displayName"
          value={state.displayName}
          aria-required
          aria-invalid={!!errors.displayName}
          className={fieldErrorClass(!!errors.displayName)}
          onChange={(e) => onChange({ displayName: e.target.value })}
          onBlur={() => onFieldBlur?.('displayName')}
        />
      </FormField>
      <div className="grid grid-cols-2 gap-2">
        <FormField
          id="host"
          label={t('connectionForm.fields.host')}
          error={formatFieldError(t, 'host', errors.host)}
        >
          <Input
            id="host"
            value={my.host}
            aria-invalid={!!errors.host}
            className={fieldErrorClass(!!errors.host)}
            onChange={(e) => setMySQL({ host: e.target.value })}
            onBlur={() => onFieldBlur?.('host')}
          />
        </FormField>
        <FormField
          id="port"
          label={t('connectionForm.fields.port')}
          error={formatFieldError(t, 'port', errors.port)}
        >
          <Input
            id="port"
            type="number"
            value={my.port}
            aria-invalid={!!errors.port}
            className={fieldErrorClass(!!errors.port)}
            onChange={(e) => setMySQL({ port: Number(e.target.value) || my.port })}
            onBlur={() => onFieldBlur?.('port')}
          />
        </FormField>
      </div>
      <FormField id="database" label={t('connectionForm.fields.database')} optional={optionalLabel}>
        <Input
          id="database"
          value={my.database}
          onChange={(e) => setMySQL({ database: e.target.value })}
          onBlur={() => onFieldBlur?.('database')}
        />
      </FormField>
      <FormField
        id="user"
        label={t('connectionForm.fields.user')}
        required
        error={formatFieldError(t, 'user', errors.user)}
      >
        <Input
          id="user"
          value={my.user}
          aria-required
          aria-invalid={!!errors.user}
          className={fieldErrorClass(!!errors.user)}
          onChange={(e) => setMySQL({ user: e.target.value })}
          onBlur={() => onFieldBlur?.('user')}
        />
      </FormField>
      <FormField
        id="password"
        label={t('connectionForm.fields.password')}
        required={mode === 'create'}
        error={formatFieldError(t, 'password', errors.password)}
      >
        <Input
          id="password"
          type="password"
          value={state.password}
          placeholder={mode === 'edit' ? t('connectionForm.passwordKeep') : undefined}
          aria-required={mode === 'create'}
          aria-invalid={!!errors.password}
          className={fieldErrorClass(!!errors.password)}
          onChange={(e) => onChange({ password: e.target.value })}
          onBlur={() => onFieldBlur?.('password')}
        />
        {mode === 'edit' && (
          <p className="text-xs text-muted">{t('connectionForm.passwordKeep')}</p>
        )}
      </FormField>
    </div>
  )
}

export function MySQLSecurityFields({
  state,
  onChange,
}: {
  state: ConnectionFormState
  onChange: (patch: Partial<ConnectionFormState>) => void
}) {
  const { t } = useTranslation()
  const my = state.mysql
  const setMySQL = (patch: Partial<typeof my>) => onChange({ mysql: { ...my, ...patch } })

  return (
    <div className="space-y-3">
      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={my.tls}
          onChange={(e) =>
            setMySQL({
              tls: e.target.checked,
              tlsSkipVerify: e.target.checked ? my.tlsSkipVerify : false,
            })
          }
        />
        {t('connection.tls')}
      </label>
      {my.tls && (
        <>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={my.tlsSkipVerify ?? false}
              onChange={(e) => setMySQL({ tlsSkipVerify: e.target.checked })}
            />
            {t('connection.tlsSkipVerify')}
          </label>
          <p className="text-xs text-muted">{t('connection.tlsSkipVerifyHint')}</p>
        </>
      )}
      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={my.readOnly}
          onChange={(e) => setMySQL({ readOnly: e.target.checked })}
        />
        {t('connection.readOnly')}
      </label>
    </div>
  )
}

export function MySQLAdvancedFields({
  state,
  onChange,
}: {
  state: ConnectionFormState
  onChange: (patch: Partial<ConnectionFormState>) => void
}) {
  const { t } = useTranslation()
  const my = state.mysql
  const setMySQL = (patch: Partial<typeof my>) => onChange({ mysql: { ...my, ...patch } })

  const charset = my.charset ?? 'utf8mb4'
  const collations = MYSQL_COLLATIONS[charset] ?? MYSQL_COLLATIONS.utf8mb4

  return (
    <div className="space-y-3">
      <FormField id="charset" label={t('connectionForm.charset')}>
        <select
          id="charset"
          className="w-full rounded border border-border bg-transparent px-3 py-2 text-sm"
          value={charset}
          onChange={(e) => {
            const nextCharset = e.target.value
            const nextCollations = MYSQL_COLLATIONS[nextCharset] ?? MYSQL_COLLATIONS.utf8mb4
            setMySQL({
              charset: nextCharset,
              collation: nextCollations[0],
            })
          }}
        >
          {(['utf8mb4', 'utf8', 'latin1', 'gbk'] as const).map((cs) => (
            <option key={cs} value={cs}>
              {cs}
            </option>
          ))}
        </select>
      </FormField>
      <FormField id="collation" label={t('connectionForm.collation')}>
        <select
          id="collation"
          className="w-full rounded border border-border bg-transparent px-3 py-2 text-sm"
          value={my.collation ?? collations[0]}
          onChange={(e) => setMySQL({ collation: e.target.value })}
        >
          {collations.map((col) => (
            <option key={col} value={col}>
              {col}
            </option>
          ))}
        </select>
      </FormField>
      <FormField id="storageEngine" label={t('connectionForm.storageEngine')}>
        <select
          id="storageEngine"
          className="w-full rounded border border-border bg-transparent px-3 py-2 text-sm"
          value={my.defaultStorageEngine ?? ''}
          onChange={(e) => setMySQL({ defaultStorageEngine: e.target.value })}
        >
          <option value="">{t('connectionForm.storageEngineDefault')}</option>
          <option value="InnoDB">InnoDB</option>
          <option value="MyISAM">MyISAM</option>
        </select>
      </FormField>
    </div>
  )
}
