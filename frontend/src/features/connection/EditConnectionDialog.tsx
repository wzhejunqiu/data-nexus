import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'
import { useToastStore } from '@/components/ui/Toast'
import { connectionApi } from '@/lib/api/connection'
import { formatError } from '@/lib/api/errors'
import { isVaultLockedError } from '@/lib/api/secrets'
import { connectionSubtitle } from '@/lib/connectionDisplay'
import type { ConnectionListItem } from '@/lib/types'
import { VaultDialog, type VaultDialogMode } from './VaultDialog'

function EditConnectionForm({
  item,
  open,
  onOpenChange,
  onSaved,
}: {
  item: ConnectionListItem
  open: boolean
  onOpenChange: (open: boolean) => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const pushToast = useToastStore((s) => s.push)
  const [readOnly, setReadOnly] = useState(() => item.config.sqlite?.readOnly ?? false)
  const [wal, setWal] = useState(() => item.config.sqlite?.wal ?? false)
  const [postgres, setPostgres] = useState(() => ({
    host: item.config.postgres?.host ?? '',
    port: item.config.postgres?.port ?? 5432,
    database: item.config.postgres?.database ?? '',
    user: item.config.postgres?.user ?? '',
    sslMode: item.config.postgres?.sslMode ?? 'disable',
    schema: item.config.postgres?.schema ?? 'public',
    readOnly: item.config.postgres?.readOnly ?? false,
  }))
  const [mysql, setMySQL] = useState(() => ({
    host: item.config.mysql?.host ?? '',
    port: item.config.mysql?.port ?? 3306,
    database: item.config.mysql?.database ?? '',
    user: item.config.mysql?.user ?? '',
    tls: item.config.mysql?.tls ?? false,
    readOnly: item.config.mysql?.readOnly ?? false,
  }))
  const [newPassword, setNewPassword] = useState('')
  const [vaultOpen, setVaultOpen] = useState(false)
  const [vaultMode, setVaultMode] = useState<VaultDialogMode>('unlock')
  const [vaultRetry, setVaultRetry] = useState<(() => void) | null>(null)

  const save = useMutation({
    mutationFn: async () => {
      if (item.type === 'sqlite') {
        return connectionApi.updateSQLiteSettings(item.id, {
          readOnly,
          wal: readOnly ? false : wal,
        })
      }
      if (item.type === 'postgres') {
        return connectionApi.updatePostgresSettings(item.id, {
          ...postgres,
          password: newPassword || undefined,
        })
      }
      return connectionApi.updateMySQLSettings(item.id, {
        ...mysql,
        password: newPassword || undefined,
      })
    },
    onSuccess: () => {
      onOpenChange(false)
      onSaved()
      qc.invalidateQueries({ queryKey: ['connections'] })
      if (item.status === 'open') {
        qc.invalidateQueries({ queryKey: ['tables', item.id] })
      }
    },
    onError: (err) => {
      if (isVaultLockedError(err)) {
        const code = (err as { code?: string }).code ?? ''
        setVaultMode(code === 'SECRETS_VAULT_NOT_INITIALIZED' ? 'init' : 'unlock')
        setVaultRetry(() => () => save.mutate())
        setVaultOpen(true)
        return
      }
      pushToast(formatError(t, err), 'error')
    },
  })

  const subtitle = connectionSubtitle(item)

  return (
    <>
      <Dialog
        open={open}
        onOpenChange={onOpenChange}
        title={t('connection.edit')}
        footer={
          <>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={() => save.mutate()} disabled={save.isPending}>
              {t('common.confirm')}
            </Button>
          </>
        }
      >
        <div className="space-y-3">
          <p className="truncate text-xs text-muted" title={subtitle}>
            {subtitle}
          </p>
          {item.type === 'sqlite' && (
            <>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={readOnly}
                  onChange={(e) => {
                    const next = e.target.checked
                    setReadOnly(next)
                    if (next) setWal(false)
                  }}
                />
                {t('connection.readOnly')}
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={wal}
                  disabled={readOnly}
                  onChange={(e) => setWal(e.target.checked)}
                />
                {t('connection.wal')}
              </label>
            </>
          )}
          {item.type === 'postgres' && (
            <>
              <Input
                value={postgres.host}
                onChange={(e) => setPostgres((p) => ({ ...p, host: e.target.value }))}
                placeholder={t('connection.host')}
              />
              <Input
                type="number"
                value={postgres.port}
                onChange={(e) => setPostgres((p) => ({ ...p, port: Number(e.target.value) }))}
                placeholder={t('connection.port')}
              />
              <Input
                value={postgres.database}
                onChange={(e) => setPostgres((p) => ({ ...p, database: e.target.value }))}
                placeholder={t('connection.database')}
              />
              <Input
                value={postgres.user}
                onChange={(e) => setPostgres((p) => ({ ...p, user: e.target.value }))}
                placeholder={t('connection.user')}
              />
              <Input
                value={postgres.schema}
                onChange={(e) => setPostgres((p) => ({ ...p, schema: e.target.value }))}
                placeholder={t('connection.schema')}
              />
              <Input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder={t('connection.passwordOptional')}
              />
            </>
          )}
          {item.type === 'mysql' && (
            <>
              <Input
                value={mysql.host}
                onChange={(e) => setMySQL((m) => ({ ...m, host: e.target.value }))}
                placeholder={t('connection.host')}
              />
              <Input
                type="number"
                value={mysql.port}
                onChange={(e) => setMySQL((m) => ({ ...m, port: Number(e.target.value) }))}
                placeholder={t('connection.port')}
              />
              <Input
                value={mysql.database}
                onChange={(e) => setMySQL((m) => ({ ...m, database: e.target.value }))}
                placeholder={t('connection.database')}
              />
              <Input
                value={mysql.user}
                onChange={(e) => setMySQL((m) => ({ ...m, user: e.target.value }))}
                placeholder={t('connection.user')}
              />
              <Input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder={t('connection.passwordOptional')}
              />
            </>
          )}
        </div>
      </Dialog>
      <VaultDialog
        open={vaultOpen}
        mode={vaultMode}
        onOpenChange={setVaultOpen}
        onSuccess={() => vaultRetry?.()}
      />
    </>
  )
}

export function EditConnectionDialog({
  item,
  open,
  onOpenChange,
  onSaved,
}: {
  item: ConnectionListItem | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSaved: () => void
}) {
  if (!item || !open) return null

  return (
    <EditConnectionForm
      key={item.id}
      item={item}
      open={open}
      onOpenChange={onOpenChange}
      onSaved={onSaved}
    />
  )
}
