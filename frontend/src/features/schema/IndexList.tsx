import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import type { IndexInfo } from '@/lib/types'

export function IndexList({ indexes }: { indexes: IndexInfo[] | null | undefined }) {
  const { t } = useTranslation()
  if (!indexes?.length) return null
  return (
    <div>
      <h4 className="mb-2 text-sm font-semibold">{t('schema.indexes')}</h4>
      <ul className="space-y-1 text-sm">
        {indexes.map((idx) => (
          <li key={idx.name} className="font-mono text-xs">
            {idx.name} ({idx.columns.join(', ')}) {idx.unique ? '[unique]' : ''}
          </li>
        ))}
      </ul>
    </div>
  )
}

export function RenameDialog({
  open,
  name,
  onNameChange,
  onOpenChange,
  onConfirm,
}: {
  open: boolean
  name: string
  onNameChange: (name: string) => void
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}) {
  const { t } = useTranslation()
  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={t('connection.rename')}
      footer={
        <>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={onConfirm}>{t('common.confirm')}</Button>
        </>
      }
    >
      <input
        className="w-full rounded border border-border bg-transparent px-3 py-2 text-sm"
        value={name}
        onChange={(e) => onNameChange(e.target.value)}
      />
    </Dialog>
  )
}
