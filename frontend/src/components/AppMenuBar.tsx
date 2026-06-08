import { useTranslation } from 'react-i18next'
import { useCallback } from 'react'
import { Quit } from '../../wailsjs/runtime/runtime'
import { connectionApi } from '@/lib/api/connection'
import { dialogApi } from '@/lib/api/dialog'
import { useIsMacOS } from '@/lib/platform'
import {
  Menubar,
  MenubarContent,
  MenubarItem,
  MenubarMenu,
  MenubarSeparator,
  MenubarShortcut,
  MenubarTrigger,
} from '@/components/ui/Menubar'

export interface AppMenuBarProps {
  openCount: number
  wizardTitle?: string
  activeConnectionId: string | null
  onNewConnection: () => void
  onNewGroup: () => void
  onOpenSettings: () => void
  onOpenSqlHistory: () => void
  onOpenAbout: () => void
  disabled?: boolean
}

function modKey(e: React.KeyboardEvent | KeyboardEvent) {
  return e.metaKey || e.ctrlKey
}

export function useAppShortcuts(props: AppMenuBarProps) {
  const { activeConnectionId, onNewConnection, onOpenSettings, disabled } = props

  const handleKeyDown = useCallback(
    async (e: KeyboardEvent) => {
      if (disabled || !modKey(e)) return
      const key = e.key.toLowerCase()
      if (key === 'n') {
        e.preventDefault()
        onNewConnection()
      } else if (key === 'o') {
        e.preventDefault()
        try {
          const path = await dialogApi.openDatabaseFile()
          if (path)
            await connectionApi.openFromFile({ filePath: path, readOnly: false, wal: false })
        } catch {
          /* cancelled or error via toast */
        }
      } else if (key === 'w') {
        if (!activeConnectionId) return
        e.preventDefault()
        await connectionApi.close(activeConnectionId)
      } else if (key === ',') {
        e.preventDefault()
        onOpenSettings()
      }
    },
    [activeConnectionId, disabled, onNewConnection, onOpenSettings],
  )

  return { handleKeyDown }
}

export function AppMenuBar({
  openCount,
  wizardTitle,
  activeConnectionId,
  onNewConnection,
  onNewGroup,
  onOpenSettings,
  onOpenSqlHistory,
  onOpenAbout,
  disabled,
}: AppMenuBarProps) {
  const { t } = useTranslation()
  const isMacOS = useIsMacOS()
  const mod = typeof navigator !== 'undefined' && /Mac/.test(navigator.platform) ? '⌘' : 'Ctrl+'

  const openSqlite = async () => {
    try {
      const path = await dialogApi.openDatabaseFile()
      if (path) await connectionApi.openFromFile({ filePath: path, readOnly: false, wal: false })
    } catch {
      /* handled globally */
    }
  }

  const closeActive = async () => {
    if (activeConnectionId) await connectionApi.close(activeConnectionId)
  }

  return (
    <div
      className={`flex w-full items-center gap-4 ${isMacOS ? 'justify-end' : 'justify-between'}`}
    >
      {!isMacOS && (
        <Menubar>
          <MenubarMenu>
            <MenubarTrigger>{t('menu.file.label')}</MenubarTrigger>
            <MenubarContent>
              <MenubarItem disabled={disabled} onSelect={onNewConnection}>
                {t('menu.file.newConnection')}
                <MenubarShortcut>{mod}N</MenubarShortcut>
              </MenubarItem>
              <MenubarItem disabled={disabled} onSelect={() => void openSqlite()}>
                {t('menu.file.openSqlite')}
                <MenubarShortcut>{mod}O</MenubarShortcut>
              </MenubarItem>
              <MenubarItem
                disabled={disabled || !activeConnectionId}
                onSelect={() => void closeActive()}
              >
                {t('menu.file.closeConnection')}
                <MenubarShortcut>{mod}W</MenubarShortcut>
              </MenubarItem>
              <MenubarItem disabled={disabled} onSelect={onNewGroup}>
                {t('menu.file.newGroup')}
              </MenubarItem>
              <MenubarSeparator />
              <MenubarItem onSelect={() => Quit()}>{t('menu.file.quit')}</MenubarItem>
            </MenubarContent>
          </MenubarMenu>
          <MenubarMenu>
            <MenubarTrigger>{t('menu.view.label')}</MenubarTrigger>
            <MenubarContent>
              <MenubarItem disabled={disabled} onSelect={onOpenSqlHistory}>
                {t('menu.view.sqlHistory')}
              </MenubarItem>
              <MenubarItem disabled={disabled} onSelect={onOpenSettings}>
                {t('menu.view.settings')}
                <MenubarShortcut>{mod},</MenubarShortcut>
              </MenubarItem>
            </MenubarContent>
          </MenubarMenu>
          <MenubarMenu>
            <MenubarTrigger>{t('menu.help.label')}</MenubarTrigger>
            <MenubarContent>
              <MenubarItem disabled={disabled} onSelect={onOpenAbout}>
                {t('menu.help.about')}
              </MenubarItem>
            </MenubarContent>
          </MenubarMenu>
        </Menubar>
      )}
      <span className="text-sm text-muted">
        {wizardTitle ?? t('app.openCount', { count: openCount })}
      </span>
    </div>
  )
}
