import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { EventsOn } from '../../../../wailsjs/runtime/runtime'
import { connectionApi } from '@/lib/api/connection'
import { formatError } from '@/lib/api/errors'
import { useToastStore } from '@/components/ui/Toast'
import { useWorkspaceStore } from '@/stores/workspaceStore'

export type FileAttachTarget = {
  connectionId: string
  paths: string[]
}

function findSqliteDropTarget(x: number, y: number): HTMLElement | null {
  const el = document.elementFromPoint(x, y)
  return el?.closest('[data-sqlite-drop-target]') as HTMLElement | null
}

function isDbPath(path: string) {
  return /\.(db|sqlite|sqlite3)$/i.test(path)
}

function hasWailsRuntime() {
  return Boolean(
    (globalThis as typeof globalThis & { runtime?: { EventsOnMultiple?: unknown } }).runtime
      ?.EventsOnMultiple,
  )
}

export function useSidebarFileDrop({
  onAttach,
  onRefresh,
}: {
  onAttach: (target: FileAttachTarget) => void
  onRefresh: () => void
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)

  useEffect(() => {
    if (!hasWailsRuntime()) {
      return
    }
    return EventsOn('app:file-drop', (payload: { x?: number; y?: number; paths?: string[] }) => {
      const paths = (payload.paths ?? []).filter(isDbPath)
      if (paths.length === 0) return

      const x = payload.x ?? 0
      const y = payload.y ?? 0
      const target = findSqliteDropTarget(x, y)

      if (target) {
        const connectionId = target.getAttribute('data-connection-id') ?? ''
        const isOpen = target.getAttribute('data-connection-open') === 'true'
        const isSqlite = target.getAttribute('data-connection-type') === 'sqlite'
        if (!isSqlite || !isOpen) {
          pushToast(t('attach.dropRequiresOpenSqlite'), 'error')
          return
        }
        onAttach({ connectionId, paths })
        return
      }

      void (async () => {
        try {
          const conn = await connectionApi.openFromFile({
            filePath: paths[0],
            readOnly: false,
            wal: false,
          })
          useWorkspaceStore.getState().setActiveConnectionId(conn.id)
          onRefresh()
          if (paths.length > 1) {
            pushToast(t('attach.multiDropOpenedFirst'), 'info')
          }
        } catch (err) {
          pushToast(formatError(t, err), 'error')
        }
      })()
    })
  }, [onAttach, onRefresh, pushToast, t])
}
