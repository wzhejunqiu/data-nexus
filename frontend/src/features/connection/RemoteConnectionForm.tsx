import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useToastStore } from '@/components/ui/Toast'
import { connectionApi } from '@/lib/api/connection'
import type { TFunction } from 'i18next'
import { formatError } from '@/lib/api/errors'
import { isVaultLockedError } from '@/lib/api/secrets'
import type { PostgresConfig, MySQLConfig } from '@/lib/types'

const defaultPostgres = (): PostgresConfig => ({
  host: 'localhost',
  port: 5432,
  database: '',
  user: '',
  sslMode: 'disable',
  schema: 'public',
  readOnly: false,
})

const defaultMySQL = (): MySQLConfig => ({
  host: 'localhost',
  port: 3306,
  database: '',
  user: '',
  tls: false,
  tlsSkipVerify: false,
  readOnly: false,
})

export function RemoteConnectionForm({
  type,
  onSaved,
  onNeedVault,
}: {
  type: 'postgres' | 'mysql'
  onSaved: (connectionId: string) => void
  onNeedVault: (mode: 'init' | 'unlock', retry: () => void) => void
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [postgres, setPostgres] = useState(defaultPostgres)
  const [mysql, setMySQL] = useState(defaultMySQL)

  const cfg = type === 'postgres' ? postgres : mysql
  const setHost = (v: string) =>
    type === 'postgres'
      ? setPostgres((p) => ({ ...p, host: v }))
      : setMySQL((m) => ({ ...m, host: v }))
  const setPort = (v: number) =>
    type === 'postgres'
      ? setPostgres((p) => ({ ...p, port: v }))
      : setMySQL((m) => ({ ...m, port: v }))
  const setDatabase = (v: string) =>
    type === 'postgres'
      ? setPostgres((p) => ({ ...p, database: v }))
      : setMySQL((m) => ({ ...m, database: v }))
  const setUser = (v: string) =>
    type === 'postgres'
      ? setPostgres((p) => ({ ...p, user: v }))
      : setMySQL((m) => ({ ...m, user: v }))
  const setReadOnly = (v: boolean) =>
    type === 'postgres'
      ? setPostgres((p) => ({ ...p, readOnly: v }))
      : setMySQL((m) => ({ ...m, readOnly: v }))

  const test = useMutation({
    mutationFn: () =>
      connectionApi.test({
        type,
        password,
        postgres: type === 'postgres' ? postgres : undefined,
        mysql: type === 'mysql' ? mysql : undefined,
      }),
    onSuccess: () => pushToast(t('connection.testSuccess'), 'success'),
    onError: (err) => pushToast(formatConnectionError(t, err), 'error'),
  })

  const save = useMutation({
    mutationFn: () =>
      connectionApi.createRemote({
        type,
        name: name.trim(),
        password,
        open: true,
        postgres: type === 'postgres' ? postgres : undefined,
        mysql: type === 'mysql' ? mysql : undefined,
      }),
    onSuccess: (saved) => {
      pushToast(t('connection.saveSuccess'), 'success')
      onSaved(saved.id)
    },
    onError: (err) => {
      if (isVaultLockedError(err)) {
        const code = (err as { code?: string }).code ?? ''
        onNeedVault(code === 'SECRETS_VAULT_NOT_INITIALIZED' ? 'init' : 'unlock', () =>
          save.mutate(),
        )
        return
      }
      pushToast(formatConnectionError(t, err), 'error')
    },
  })

  return (
    <div className="space-y-3">
      <Input
        placeholder={t('connection.displayName')}
        value={name}
        onChange={(e) => setName(e.target.value)}
      />
      <div className="grid grid-cols-2 gap-2">
        <Input
          placeholder={t('connection.host')}
          value={cfg.host}
          onChange={(e) => setHost(e.target.value)}
        />
        <Input
          type="number"
          placeholder={t('connection.port')}
          value={cfg.port}
          onChange={(e) => setPort(Number(e.target.value) || cfg.port)}
        />
      </div>
      <Input
        placeholder={t('connection.database')}
        value={cfg.database}
        onChange={(e) => setDatabase(e.target.value)}
      />
      <Input
        placeholder={t('connection.user')}
        value={cfg.user}
        onChange={(e) => setUser(e.target.value)}
      />
      <Input
        type="password"
        placeholder={t('connection.password')}
        value={password}
        onChange={(e) => setPassword(e.target.value)}
      />
      {type === 'postgres' && (
        <>
          <Input
            placeholder={t('connection.schema')}
            value={postgres.schema ?? 'public'}
            onChange={(e) => setPostgres((p) => ({ ...p, schema: e.target.value }))}
          />
          <label className="flex items-center gap-2 text-xs text-muted">
            <select
              className="rounded border border-border bg-transparent px-2 py-1 text-xs"
              value={postgres.sslMode ?? 'disable'}
              onChange={(e) => setPostgres((p) => ({ ...p, sslMode: e.target.value }))}
            >
              <option value="disable">SSL disable</option>
              <option value="require">SSL require</option>
              <option value="verify-full">SSL verify-full</option>
            </select>
          </label>
        </>
      )}
      {type === 'mysql' && (
        <>
          <label className="flex items-center gap-2 text-xs text-muted">
            <input
              type="checkbox"
              checked={mysql.tls}
              onChange={(e) =>
                setMySQL((m) => ({
                  ...m,
                  tls: e.target.checked,
                  tlsSkipVerify: e.target.checked ? m.tlsSkipVerify : false,
                }))
              }
            />
            {t('connection.tls')}
          </label>
          {mysql.tls && (
            <label className="flex items-center gap-2 text-xs text-muted">
              <input
                type="checkbox"
                checked={mysql.tlsSkipVerify ?? false}
                onChange={(e) => setMySQL((m) => ({ ...m, tlsSkipVerify: e.target.checked }))}
              />
              {t('connection.tlsSkipVerify')}
            </label>
          )}
          {mysql.tls && <p className="text-xs text-muted">{t('connection.tlsSkipVerifyHint')}</p>}
        </>
      )}
      <label className="flex items-center gap-2 text-xs text-muted">
        <input
          type="checkbox"
          checked={cfg.readOnly}
          onChange={(e) => setReadOnly(e.target.checked)}
        />
        {t('connection.readOnly')}
      </label>
      <div className="flex gap-2">
        <Button variant="outline" size="sm" onClick={() => test.mutate()} disabled={test.isPending}>
          {t('connection.test')}
        </Button>
        <Button size="sm" onClick={() => save.mutate()} disabled={save.isPending}>
          {t('connection.saveAndOpen')}
        </Button>
      </div>
    </div>
  )
}

function formatConnectionError(t: TFunction, err: unknown): string {
  if (
    err &&
    typeof err === 'object' &&
    'code' in err &&
    (err as { code: string }).code === 'CONNECTION_FAILED'
  ) {
    const details = (err as { details?: { reason?: string } }).details
    const reason = details?.reason
    if (reason === 'auth') return t('connection.error.auth')
    if (reason === 'network') return t('connection.error.network')
    if (reason === 'database') return t('connection.error.database')
  }
  return formatError(t, err)
}
