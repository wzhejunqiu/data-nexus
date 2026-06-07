import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { Button } from './ui/Button'
import { SettingsDialogContainer } from '@/features/settings/SettingsDialog'
import { useThemeStore, resolveTheme, type ThemeMode } from '@/stores/themeStore'

export function Header({ openCount }: { openCount: number }) {
  const { t, i18n } = useTranslation()
  const { mode, setMode } = useThemeStore()
  const [settingsOpen, setSettingsOpen] = useState(false)

  useEffect(() => {
    return EventsOn('app:settings', () => setSettingsOpen(true))
  }, [])

  const cycleTheme = () => {
    const order: ThemeMode[] = ['light', 'dark', 'system']
    const idx = order.indexOf(mode)
    setMode(order[(idx + 1) % order.length])
    applyTheme(order[(idx + 1) % order.length])
  }

  const applyTheme = (m: ThemeMode) => {
    const resolved = resolveTheme(m)
    document.documentElement.classList.toggle('dark', resolved === 'dark')
  }

  const toggleLang = () => {
    const next = i18n.language === 'zh-CN' ? 'en' : 'zh-CN'
    i18n.changeLanguage(next)
    localStorage.setItem('data-nexus-lang', next)
  }

  return (
    <header className="flex h-12 items-center justify-between border-b border-border px-4">
      <div className="flex items-center gap-3">
        <span className="font-semibold">{t('app.title')}</span>
        <span className="text-sm text-muted">{t('app.openCount', { count: openCount })}</span>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" onClick={toggleLang}>
          {i18n.language === 'zh-CN' ? t('language.en') : t('language.zh')}
        </Button>
        <Button variant="ghost" size="sm" onClick={() => setSettingsOpen(true)}>
          {t('settings.title')}
        </Button>
        <Button variant="ghost" size="sm" onClick={cycleTheme}>
          {t(`theme.${mode}`)}
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
