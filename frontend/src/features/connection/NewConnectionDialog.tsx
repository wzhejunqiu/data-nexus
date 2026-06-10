import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useCallback, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { useToastStore } from '@/components/ui/Toast'
import { connectionApi } from '@/lib/api/connection'
import { dialogApi } from '@/lib/api/dialog'
import { formatError, isDialogCancelled, mapWailsError } from '@/lib/api/errors'
import { isVaultLockedError } from '@/lib/api/secrets'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import { handleRemoteConnectionError } from './connectionFormFailure'
import { moveNewConnectionToGroup, openSqliteAndMaybePlace } from './placeConnectionInGroup'
import { VaultDialog, type VaultDialogMode } from './VaultDialog'
import {
  ConnectionForm,
  type ConnectionFormHandle,
  type ConnectionFormState,
} from './ConnectionForm/ConnectionForm'

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
  const selectedGroupId = useWorkspaceStore((s) => s.selectedGroupId)
  const formRef = useRef<ConnectionFormHandle>(null)
  const [vaultOpen, setVaultOpen] = useState(false)
  const [vaultMode, setVaultMode] = useState<VaultDialogMode>('unlock')
  const [vaultRetry, setVaultRetry] = useState<(() => void) | null>(null)

  const onNeedVault = useCallback((mode: VaultDialogMode, retry: () => void) => {
    setVaultMode(mode)
    setVaultRetry(() => retry)
    setVaultOpen(true)
  }, [])

  const createSQLite = useMutation({
    mutationFn: async (state: ConnectionFormState) => {
      let path = state.filePath.trim()
      if (!path) {
        path = (await dialogApi.openDatabaseFile()) ?? ''
      }
      if (!path) throw new Error('cancelled')
      return openSqliteAndMaybePlace(
        {
          filePath: path,
          readOnly: state.readOnly,
          wal: state.readOnly ? false : state.wal,
        },
        selectedGroupId,
        qc,
      )
    },
    onSuccess: (conn) => {
      onOpenChange(false)
      setActiveConnectionId(conn.id)
      qc.invalidateQueries({ queryKey: ['connections'] })
      qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
      qc.invalidateQueries({ queryKey: ['tables', conn.id] })
    },
    onError: (err) => {
      const appErr = mapWailsError(err)
      if (isDialogCancelled(appErr)) return
      pushToast(formatError(t, appErr), 'error')
    },
  })

  const createRemote = useMutation({
    mutationFn: (state: ConnectionFormState) =>
      connectionApi.createRemote({
        type: state.dialect as 'postgres' | 'mysql',
        name: state.displayName.trim(),
        password: state.password,
        open: true,
        postgres: state.dialect === 'postgres' ? state.postgres : undefined,
        mysql: state.dialect === 'mysql' ? state.mysql : undefined,
      }),
    onSuccess: async (saved) => {
      await moveNewConnectionToGroup(saved.id, selectedGroupId, qc)
      pushToast(t('connection.saveSuccess'), 'success')
      onOpenChange(false)
      setActiveConnectionId(saved.id)
      qc.invalidateQueries({ queryKey: ['connections'] })
      qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
      qc.invalidateQueries({ queryKey: ['tables', saved.id] })
    },
    onError: (err) => {
      if (isVaultLockedError(err)) {
        const code = (err as { code?: string }).code ?? ''
        onNeedVault(code === 'SECRETS_VAULT_NOT_INITIALIZED' ? 'init' : 'unlock', () =>
          formRef.current?.submit(),
        )
        return
      }
      if (handleRemoteConnectionError(err, { t, pushToast, formRef })) return
      pushToast(formatError(t, err), 'error')
    },
  })

  const testRemote = useMutation({
    mutationFn: (state: ConnectionFormState) =>
      connectionApi.test({
        type: state.dialect as 'postgres' | 'mysql',
        password: state.password,
        postgres: state.dialect === 'postgres' ? state.postgres : undefined,
        mysql: state.dialect === 'mysql' ? state.mysql : undefined,
      }),
    onSuccess: () => pushToast(t('connection.testSuccess'), 'success'),
    onError: (err) => {
      if (handleRemoteConnectionError(err, { t, pushToast, formRef })) return
      pushToast(formatError(t, err), 'error')
    },
  })

  const handleSubmit = async (state: ConnectionFormState) => {
    if (state.dialect === 'sqlite') {
      let path = state.filePath.trim()
      if (!path) {
        try {
          path = await dialogApi.openDatabaseFile()
        } catch (err) {
          const appErr = mapWailsError(err)
          if (!isDialogCancelled(appErr)) {
            pushToast(formatError(t, appErr), 'error')
          }
          return
        }
      }
      createSQLite.mutate({ ...state, filePath: path })
      return
    }
    createRemote.mutate(state)
  }

  const handleBrowse = async (): Promise<string | undefined> => {
    try {
      return await dialogApi.openDatabaseFile()
    } catch (err) {
      const appErr = mapWailsError(err)
      if (!isDialogCancelled(appErr)) {
        pushToast(formatError(t, appErr), 'error')
      }
      return undefined
    }
  }

  const pending = createSQLite.isPending || createRemote.isPending

  return (
    <>
      <Dialog
        wide
        open={open}
        onOpenChange={onOpenChange}
        title={t('connectionForm.titleNew')}
        footer={
          <>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={() => formRef.current?.submit()} disabled={pending}>
              {t('connectionForm.saveAndOpen')}
            </Button>
          </>
        }
      >
        <ConnectionForm
          ref={formRef}
          mode="create"
          readOnlyDefault={readOnly}
          walDefault={wal}
          onBrowse={handleBrowse}
          onSubmit={(state) => void handleSubmit(state)}
          onTest={(state) => testRemote.mutate(state)}
          showTest
          testPending={testRemote.isPending}
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
