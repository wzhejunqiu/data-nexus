import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { Button } from './ui/Button'
import { SettingsDialogContainer } from '@/features/settings/SettingsDialog'

export function Header({ openCount, wizardTitle }: { openCount: number; wizardTitle?: string }) {
  const { t } = useTranslation()
  const [settingsOpen, setSettingsOpen] = useState(false)

  useEffect(() => {
    return EventsOn('app:settings', () => setSettingsOpen(true))
  }, [])

  return (
    <header className="flex h-12 items-center justify-between border-b border-border px-4">
      <div className="flex items-center gap-3">
        <span className="font-semibold">{t('app.title')}</span>
        {wizardTitle ? (
          <span className="text-sm text-muted">{wizardTitle}</span>
        ) : (
          <span className="text-sm text-muted">{t('app.openCount', { count: openCount })}</span>
        )}
      </div>
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" onClick={() => setSettingsOpen(true)}>
          {t('settings.title')}
        </Button>
      </div>
      <SettingsDialogContainer open={settingsOpen} onOpenChange={setSettingsOpen} />
    </header>
  )
}

export function StatusBar({ text }: { text: string }) {
  return (
    <footer className="flex h-7 items-center border-t border-border px-3 text-xs text-muted">
      {text}
    </footer>
  )
}
