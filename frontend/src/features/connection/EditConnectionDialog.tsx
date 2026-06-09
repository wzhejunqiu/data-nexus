import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useCallback, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { useToastStore } from '@/components/ui/Toast'
import { connectionApi } from '@/lib/api/connection'
import { formatError } from '@/lib/api/errors'
import { isVaultLockedError } from '@/lib/api/secrets'
import type { ConnectionListItem } from '@/lib/types'
import {
  ConnectionForm,
  type ConnectionFormHandle,
  type ConnectionFormState,
} from './ConnectionForm/ConnectionForm'
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
  const formRef = useRef<ConnectionFormHandle>(null)
  const [vaultOpen, setVaultOpen] = useState(false)
  const [vaultMode, setVaultMode] = useState<VaultDialogMode>('unlock')
  const [vaultRetry, setVaultRetry] = useState<(() => void) | null>(null)

  const save = useMutation({
    mutationFn: async (state: ConnectionFormState) => {
      if (state.displayName.trim() !== item.name) {
        await connectionApi.rename(item.id, state.displayName.trim())
      }
      if (item.type === 'sqlite') {
        return connectionApi.updateSQLiteSettings(item.id, {
          readOnly: state.readOnly,
          wal: state.readOnly ? false : state.wal,
        })
      }
      if (item.type === 'postgres') {
        return connectionApi.updatePostgresSettings(item.id, {
          host: state.postgres.host,
          port: state.postgres.port,
          database: state.postgres.database,
          user: state.postgres.user,
          sslMode: state.postgres.sslMode ?? 'disable',
          schema: state.postgres.schema ?? 'public',
          clientEncoding: state.postgres.clientEncoding ?? 'UTF8',
          readOnly: state.postgres.readOnly,
          password: state.password || undefined,
        })
      }
      return connectionApi.updateMySQLSettings(item.id, {
        host: state.mysql.host,
        port: state.mysql.port,
        database: state.mysql.database,
        user: state.mysql.user,
        tls: state.mysql.tls,
        tlsSkipVerify: state.mysql.tlsSkipVerify ?? false,
        charset: state.mysql.charset ?? 'utf8mb4',
        collation: state.mysql.collation ?? 'utf8mb4_unicode_ci',
        defaultStorageEngine: state.mysql.defaultStorageEngine ?? '',
        readOnly: state.mysql.readOnly,
        password: state.password || undefined,
      })
    },
    onSuccess: () => {
      onOpenChange(false)
      onSaved()
      qc.invalidateQueries({ queryKey: ['connections'] })
    },
    onError: (err) => {
      if (isVaultLockedError(err)) {
        const code = (err as { code?: string }).code ?? ''
        setVaultMode(code === 'SECRETS_VAULT_NOT_INITIALIZED' ? 'init' : 'unlock')
        setVaultRetry(() => () => formRef.current?.submit())
        setVaultOpen(true)
        return
      }
      pushToast(formatError(t, err), 'error')
    },
  })

  const handleSubmit = useCallback((state: ConnectionFormState) => save.mutate(state), [save])

  return (
    <>
      <Dialog
        wide
        open={open}
        onOpenChange={onOpenChange}
        title={t('connectionForm.titleEdit')}
        footer={
          <>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={() => formRef.current?.submit()} disabled={save.isPending}>
              {t('common.confirm')}
            </Button>
          </>
        }
      >
        <ConnectionForm
          ref={formRef}
          mode="edit"
          initialItem={item}
          dialectLocked
          onSubmit={handleSubmit}
        />
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
