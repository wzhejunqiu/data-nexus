import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import type { PendingEdit } from '@/lib/types/edit'
import { formatCell, pkSummary } from '@/lib/utils'

export function BatchEditConfirmDialog({
  open,
  edits,
  onOpenChange,
  onConfirm,
  busy,
}: {
  open: boolean
  edits: PendingEdit[]
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
  busy?: boolean
}) {
  const { t } = useTranslation()
  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={t('edit.batchConfirmTitle', { count: edits.length })}
      footer={
        <>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={busy}>
            {t('common.cancel')}
          </Button>
            <Button variant="danger" onClick={onConfirm} disabled={busy}>
            {t('common.confirm')}
          </Button>
        </>
      }
    >
      <div className="max-h-64 overflow-auto">
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-border text-left">
              <th className="px-2 py-1">{t('edit.row')}</th>
              <th className="px-2 py-1">{t('edit.column')}</th>
              <th className="px-2 py-1">{t('edit.change')}</th>
            </tr>
          </thead>
          <tbody>
            {edits.map((e, i) => (
              <tr key={i} className="border-b border-border/40">
                <td className="px-2 py-1 font-mono">{pkSummary(e.primaryKey)}</td>
                <td className="px-2 py-1">{e.columnName}</td>
                <td className="px-2 py-1 font-mono">
                  {formatCell(e.originalValue)} → {formatCell(e.newValue)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Dialog>
  )
}
