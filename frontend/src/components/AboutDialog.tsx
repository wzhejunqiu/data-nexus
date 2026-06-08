import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { appApi } from '@/lib/api/app'
import { Dialog } from '@/components/ui/Dialog'

export function AboutDialog({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (v: boolean) => void
}) {
  const { t } = useTranslation()
  const { data: version } = useQuery({
    queryKey: ['version'],
    queryFn: () => appApi.getVersion(),
    enabled: open,
  })
  const { data: platform } = useQuery({
    queryKey: ['platform'],
    queryFn: () => appApi.getPlatform(),
    enabled: open,
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange} title={t('about.title')}>
      <div className="space-y-2 text-sm">
        <p className="font-semibold">Data Nexus</p>
        <p>{t('about.version', { version: version?.version ?? '—' })}</p>
        <p>{t('about.platform', { platform: platform ?? version?.platform ?? '—' })}</p>
        <p className="text-muted">{t('about.description')}</p>
      </div>
    </Dialog>
  )
}
