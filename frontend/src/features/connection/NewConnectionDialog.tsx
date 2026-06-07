import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'
import { useToastStore } from '@/components/ui/Toast'
import { connectionApi } from '@/lib/api/connection'
import { dialogApi } from '@/lib/api/dialog'
import { formatError, isDialogCancelled, mapWailsError } from '@/lib/api/errors'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import { RemoteConnectionForm } from './RemoteConnectionForm'
import { VaultDialog, type VaultDialogMode } from './VaultDialog'

type ConnTab = 'sqlite' | 'postgres' | 'mysql'

export function NewConnectionDialog({
  open,
  onOpenChange,
  readOnly,
  wal,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  readOnly: boolean
  wal: boolean
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const pushToast = useToastStore((s) => s.push)
  const setActiveConnectionId = useWorkspaceStore((s) => s.setActiveConnectionId)
  const [tab, setTab] = useState<ConnTab>('sqlite')
  const [filePath, setFilePath] = useState('')
  const [vaultOpen, setVaultOpen] = useState(false)
  const [vaultMode, setVaultMode] = useState<VaultDialogMode>('unlock')
  const [vaultRetry, setVaultRetry] = useState<(() => void) | null>(null)

  const onNeedVault = useCallback((mode: VaultDialogMode, retry: () => void) => {
    setVaultMode(mode)
    setVaultRetry(() => retry)
    setVaultOpen(true)
  }, [])

  const create = useMutation({
    mutationFn: async () => {
      const path = filePath.trim() || (await dialogApi.openDatabaseFile())
      return connectionApi.openFromFile({ filePath: path, readOnly, wal: readOnly ? false : wal })
    },
    onSuccess: (conn) => {
      setFilePath('')
      onOpenChange(false)
      setActiveConnectionId(conn.id)
      qc.invalidateQueries({ queryKey: ['connections'] })
      qc.invalidateQueries({ queryKey: ['tables', conn.id] })
    },
    onError: (err) => {
      const appErr = mapWailsError(err)
      if (isDialogCancelled(appErr)) return
      pushToast(formatError(t, appErr), 'error')
    },
  })

  const onRemoteSaved = (connectionId: string) => {
    onOpenChange(false)
    setActiveConnectionId(connectionId)
    qc.invalidateQueries({ queryKey: ['connections'] })
    qc.invalidateQueries({ queryKey: ['tables', connectionId] })
  }

  return (
    <>
      <Dialog
        open={open}
        onOpenChange={onOpenChange}
        title={t('connection.new')}
        footer={
          tab === 'sqlite' ? (
            <>
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                {t('common.cancel')}
              </Button>
              <Button onClick={() => create.mutate()} disabled={create.isPending}>
                {t('connection.open')}
              </Button>
            </>
          ) : null
        }
      >
        <div className="mb-3 flex gap-1">
          {(['sqlite', 'postgres', 'mysql'] as ConnTab[]).map((key) => (
            <Button
              key={key}
              size="sm"
              variant={tab === key ? 'default' : 'outline'}
              onClick={() => setTab(key)}
            >
              {t(`connection.type.${key}`)}
            </Button>
          ))}
        </div>
        {tab === 'sqlite' && (
          <div className="space-y-3">
            <label className="block text-xs text-muted">{t('connection.pathHint')}</label>
            <Input
              placeholder="/path/to/database.db"
              value={filePath}
              onChange={(e) => setFilePath(e.target.value)}
            />
            <Button
              variant="outline"
              size="sm"
              onClick={async () => {
                try {
                  const path = await dialogApi.openDatabaseFile()
                  setFilePath(path)
                } catch (err) {
                  const appErr = mapWailsError(err)
                  if (!isDialogCancelled(appErr)) {
                    pushToast(formatError(t, appErr), 'error')
                  }
                }
              }}
            >
              {t('connection.browse')}
            </Button>
          </div>
        )}
        {tab === 'postgres' && (
          <RemoteConnectionForm type="postgres" onSaved={onRemoteSaved} onNeedVault={onNeedVault} />
        )}
        {tab === 'mysql' && (
          <RemoteConnectionForm type="mysql" onSaved={onRemoteSaved} onNeedVault={onNeedVault} />
        )}
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
