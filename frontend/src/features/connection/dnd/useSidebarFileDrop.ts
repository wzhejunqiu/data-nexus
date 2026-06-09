import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useQueryClient } from '@tanstack/react-query'
import { EventsOn } from '../../../../wailsjs/runtime/runtime'
import { formatError } from '@/lib/api/errors'
import { useToastStore } from '@/components/ui/Toast'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import { openSqliteAndMaybePlace } from '../placeConnectionInGroup'

export type FileAttachTarget = {
  connectionId: string
  paths: string[]
}

function findSqliteDropTarget(x: number, y: number): HTMLElement | null {
  const el = document.elementFromPoint(x, y)
  return el?.closest('[data-sqlite-drop-target]') as HTMLElement | null
}

function findGroupDropTarget(x: number, y: number): HTMLElement | null {
  const el = document.elementFromPoint(x, y)
  return el?.closest('[data-group-drop-target]') as HTMLElement | null
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
  const qc = useQueryClient()

  useEffect(() => {
    if (!hasWailsRuntime()) {
      return
    }
    return EventsOn('app:file-drop', (payload: { x?: number; y?: number; paths?: string[] }) => {
      const paths = (payload.paths ?? []).filter(isDbPath)
      if (paths.length === 0) return

      const x = payload.x ?? 0
      const y = payload.y ?? 0
      const sqliteTarget = findSqliteDropTarget(x, y)

      if (sqliteTarget) {
        const connectionId = sqliteTarget.getAttribute('data-connection-id') ?? ''
        const isOpen = sqliteTarget.getAttribute('data-connection-open') === 'true'
        const isSqlite = sqliteTarget.getAttribute('data-connection-type') === 'sqlite'
        if (!isSqlite || !isOpen) {
          pushToast(t('attach.dropRequiresOpenSqlite'), 'error')
          return
        }
        onAttach({ connectionId, paths })
        return
      }

      const groupTarget = findGroupDropTarget(x, y)
      const groupId = groupTarget?.getAttribute('data-group-id') ?? null

      void (async () => {
        try {
          const conn = await openSqliteAndMaybePlace(
            {
              filePath: paths[0],
              readOnly: false,
              wal: false,
            },
            groupId,
            qc,
          )
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
  }, [onAttach, onRefresh, pushToast, qc, t])
}
