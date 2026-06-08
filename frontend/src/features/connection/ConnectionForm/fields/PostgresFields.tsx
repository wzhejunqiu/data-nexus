import { useTranslation } from 'react-i18next'
import { Input } from '@/components/ui/Input'
import { fieldErrorClass, FormField } from '../FormField'
import type { ConnectionFormErrors } from '../connectionFormValidation'
import type { ConnectionFormState } from '../connectionFormDefaults'

export function PostgresFields({
  state,
  errors,
  mode,
  onChange,
}: {
  state: ConnectionFormState
  errors: ConnectionFormErrors
  mode: 'create' | 'edit'
  onChange: (patch: Partial<ConnectionFormState>) => void
}) {
  const { t } = useTranslation()
  const pg = state.postgres

  const setPostgres = (patch: Partial<typeof pg>) => onChange({ postgres: { ...pg, ...patch } })

  return (
    <div className="space-y-3">
      <FormField
        id="displayName"
        label={t('connectionForm.fields.displayName')}
        required
        error={
          errors.displayName
            ? t('connectionForm.validation.required', {
                field: t('connectionForm.fields.displayName'),
              })
            : undefined
        }
      >
        <Input
          id="displayName"
          value={state.displayName}
          aria-required
          aria-invalid={!!errors.displayName}
          className={fieldErrorClass(!!errors.displayName)}
          onChange={(e) => onChange({ displayName: e.target.value })}
        />
      </FormField>
      <div className="grid grid-cols-2 gap-2">
        <FormField
          id="host"
          label={t('connectionForm.fields.host')}
          error={
            errors.host
              ? t('connectionForm.validation.required', { field: t('connectionForm.fields.host') })
              : undefined
          }
        >
          <Input
            id="host"
            value={pg.host}
            aria-invalid={!!errors.host}
            className={fieldErrorClass(!!errors.host)}
            onChange={(e) => setPostgres({ host: e.target.value })}
          />
        </FormField>
        <FormField
          id="port"
          label={t('connectionForm.fields.port')}
          error={errors.port ? t('connectionForm.validation.invalidPort') : undefined}
        >
          <Input
            id="port"
            type="number"
            value={pg.port}
            aria-invalid={!!errors.port}
            className={fieldErrorClass(!!errors.port)}
            onChange={(e) => setPostgres({ port: Number(e.target.value) || pg.port })}
          />
        </FormField>
      </div>
      <FormField
        id="database"
        label={t('connectionForm.fields.database')}
        required
        error={
          errors.database
            ? t('connectionForm.validation.required', {
                field: t('connectionForm.fields.database'),
              })
            : undefined
        }
      >
        <Input
          id="database"
          value={pg.database}
          aria-required
          aria-invalid={!!errors.database}
          className={fieldErrorClass(!!errors.database)}
          onChange={(e) => setPostgres({ database: e.target.value })}
        />
      </FormField>
      <FormField
        id="user"
        label={t('connectionForm.fields.user')}
        required
        error={
          errors.user
            ? t('connectionForm.validation.required', { field: t('connectionForm.fields.user') })
            : undefined
        }
      >
        <Input
          id="user"
          value={pg.user}
          aria-required
          aria-invalid={!!errors.user}
          className={fieldErrorClass(!!errors.user)}
          onChange={(e) => setPostgres({ user: e.target.value })}
        />
      </FormField>
      <FormField
        id="password"
        label={t('connectionForm.fields.password')}
        required={mode === 'create'}
        error={
          errors.password
            ? t('connectionForm.validation.required', {
                field: t('connectionForm.fields.password'),
              })
            : undefined
        }
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
        />
        {mode === 'edit' && (
          <p className="text-xs text-muted">{t('connectionForm.passwordKeep')}</p>
        )}
      </FormField>
    </div>
  )
}

export function PostgresSecurityFields({
  state,
  onChange,
}: {
  state: ConnectionFormState
  onChange: (patch: Partial<ConnectionFormState>) => void
}) {
  const { t } = useTranslation()
  const pg = state.postgres
  const setPostgres = (patch: Partial<typeof pg>) => onChange({ postgres: { ...pg, ...patch } })

  return (
    <div className="space-y-3">
      <FormField id="sslMode" label={t('connectionForm.sslMode.label')}>
        <select
          id="sslMode"
          className="w-full rounded border border-border bg-transparent px-3 py-2 text-sm"
          value={pg.sslMode ?? 'disable'}
          onChange={(e) => setPostgres({ sslMode: e.target.value })}
        >
          {(['disable', 'allow', 'prefer', 'require', 'verify-ca', 'verify-full'] as const).map(
            (mode) => (
              <option key={mode} value={mode}>
                {t(`connectionForm.sslMode.${mode}`)}
              </option>
            ),
          )}
        </select>
      </FormField>
      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={pg.readOnly}
          onChange={(e) => setPostgres({ readOnly: e.target.checked })}
        />
        {t('connection.readOnly')}
      </label>
    </div>
  )
}

export function PostgresAdvancedFields({
  state,
  onChange,
}: {
  state: ConnectionFormState
  onChange: (patch: Partial<ConnectionFormState>) => void
}) {
  const { t } = useTranslation()
  const pg = state.postgres
  const setPostgres = (patch: Partial<typeof pg>) => onChange({ postgres: { ...pg, ...patch } })

  return (
    <div className="space-y-3">
      <FormField id="schema" label={t('connection.schema')}>
        <Input
          id="schema"
          value={pg.schema ?? 'public'}
          onChange={(e) => setPostgres({ schema: e.target.value })}
        />
      </FormField>
      <FormField id="clientEncoding" label={t('connectionForm.clientEncoding')}>
        <select
          id="clientEncoding"
          className="w-full rounded border border-border bg-transparent px-3 py-2 text-sm"
          value={pg.clientEncoding ?? 'UTF8'}
          onChange={(e) => setPostgres({ clientEncoding: e.target.value })}
        >
          {(['UTF8', 'LATIN1', 'SQL_ASCII', 'GBK'] as const).map((enc) => (
            <option key={enc} value={enc}>
              {enc}
            </option>
          ))}
        </select>
      </FormField>
    </div>
  )
}
