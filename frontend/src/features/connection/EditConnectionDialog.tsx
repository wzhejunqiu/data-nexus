import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { useToastStore } from '@/components/ui/Toast'
import { connectionApi } from '@/lib/api/connection'
import { formatError } from '@/lib/api/errors'
import type { ConnectionListItem } from '@/lib/types'

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

  const save = useMutation({
    mutationFn: () => connectionApi.updateReadOnly(item.id, readOnly),
    onSuccess: () => {
      onOpenChange(false)
      onSaved()
      qc.invalidateQueries({ queryKey: ['connections'] })
      if (item.status === 'open') {
        qc.invalidateQueries({ queryKey: ['tables', item.id] })
      }
    },
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

  const path = item.config.sqlite?.filePath ?? ''

  return (
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
        <p className="truncate font-mono text-xs text-muted" title={path}>
          {path}
        </p>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={readOnly}
            onChange={(e) => setReadOnly(e.target.checked)}
          />
          {t('connection.readOnly')}
        </label>
        <p className="text-xs text-muted">{t('connection.readOnlyHint')}</p>
      </div>
    </Dialog>
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
