import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'
import { useToastStore } from '@/components/ui/Toast'
import { connectionApi } from '@/lib/api/connection'
import { dialogApi } from '@/lib/api/dialog'
import { formatError, isDialogCancelled, mapWailsError } from '@/lib/api/errors'
import type { AppError } from '@/lib/types'
import { useWorkspaceStore } from '@/stores/workspaceStore'

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
  const [filePath, setFilePath] = useState('')

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
      pushToast(formatError(t, appErr as AppError), 'error')
    },
  })

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={t('connection.new')}
      footer={
        <>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={() => create.mutate()} disabled={create.isPending}>
            {t('connection.open')}
          </Button>
        </>
      }
    >
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
    </Dialog>
  )
}
