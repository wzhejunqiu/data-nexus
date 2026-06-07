import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'

export function ConfirmDialog({
  open,
  message,
  onOpenChange,
  onConfirm,
}: {
  open: boolean
  message: string
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}) {
  const { t } = useTranslation()
  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={t('common.confirm')}
      footer={
        <>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button variant="danger" onClick={onConfirm}>
            {t('common.confirm')}
          </Button>
        </>
      }
    >
      <p>{message}</p>
    </Dialog>
  )
}
