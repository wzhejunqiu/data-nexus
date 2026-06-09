import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { connectionGroupApi } from '@/lib/api/connectionGroup'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { formatError } from '@/lib/api/errors'
import { useToastStore } from '@/components/ui/Toast'

export function DeleteGroupDialog({
  groupId,
  groupName,
  open,
  onOpenChange,
  onDeleted,
}: {
  groupId: string
  groupName: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onDeleted: () => void
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const [deleteConnections, setDeleteConnections] = useState(false)
  const [busy, setBusy] = useState(false)

  const { data: preview } = useQuery({
    queryKey: ['groupDeletePreview', groupId],
    queryFn: () => connectionGroupApi.getGroupDeletePreview(groupId),
    enabled: open,
  })

  const connCount = preview?.connCount ?? 0
  const subgroupCount = preview?.subgroupCount ?? 0

  const handleOpenChange = (next: boolean) => {
    if (!next) {
      setDeleteConnections(false)
      setBusy(false)
    }
    onOpenChange(next)
  }

  const handleDelete = async () => {
    setBusy(true)
    try {
      await connectionGroupApi.deleteGroup({ id: groupId, deleteConnections })
      handleOpenChange(false)
      onDeleted()
    } catch (err) {
      pushToast(formatError(t, err), 'error')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange} title={t('connectionGroup.delete')}>
      <div className="space-y-3 text-sm">
        {connCount === 0 ? (
          <p>{t('connectionGroup.deleteSimpleConfirm', { name: groupName })}</p>
        ) : (
          <>
            <p>
              {t('connectionGroup.deleteWithConnections', {
                name: groupName,
                subgroupCount,
                connCount,
              })}
            </p>
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={deleteConnections}
                onChange={(e) => setDeleteConnections(e.target.checked)}
              />
              {t('connectionGroup.deleteConnectionsToo', { connCount })}
            </label>
            <p className="text-xs text-muted">{t('connectionGroup.deleteConnectionsHint')}</p>
          </>
        )}
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => handleOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button variant="danger" disabled={busy} onClick={() => void handleDelete()}>
            {t('connectionGroup.delete')}
          </Button>
        </div>
      </div>
    </Dialog>
  )
}
