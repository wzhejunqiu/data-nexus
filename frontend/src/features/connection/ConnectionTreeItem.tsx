import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/ContextMenu'
import { connectionApi } from '@/lib/api/connection'
import { formatError, mapWailsError } from '@/lib/api/errors'
import { isVaultLockedError } from '@/lib/api/secrets'
import { connectionSubtitle, connectionTypeLabel } from '@/lib/connectionDisplay'
import type { ConnectionListItem } from '@/lib/types'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import { EditConnectionDialog } from './EditConnectionDialog'
import type { VaultDialogMode } from './VaultDialog'
import { AttachDatabaseDialog } from './AttachDatabaseDialog'
import { ConnectionSchemaTree } from './tree/ConnectionSchemaTree'

export function ConnectionTreeItem({
  item,
  depth,
  sortContainerId,
  onRefresh,
  onVaultRequired,
}: {
  item: ConnectionListItem
  depth: number
  sortContainerId: string
  onRefresh: () => void
  onVaultRequired?: (id: string, mode: VaultDialogMode) => void
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const isOpen = item.status === 'open'
  const subtitle = connectionSubtitle(item)
  const setActiveConnectionId = useWorkspaceStore((s) => s.setActiveConnectionId)
  const selectTable = useWorkspaceStore((s) => s.selectTable)
  const [editOpen, setEditOpen] = useState(false)
  const [attachOpen, setAttachOpen] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: item.id,
    data: { type: 'connection', sortable: { containerId: sortContainerId } },
  })
  const dragStyle = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const openConn = useMutation({
    mutationFn: () => connectionApi.open(item.id),
    onSuccess: (conn) => {
      setError(null)
      setActiveConnectionId(conn.id)
      onRefresh()
    },
    onError: (err) => {
      const appErr = mapWailsError(err)
      if (isVaultLockedError(err) && onVaultRequired) {
        onVaultRequired(
          item.id,
          appErr.code === 'SECRETS_VAULT_NOT_INITIALIZED' ? 'init' : 'unlock',
        )
        return
      }
      setError(formatError(t, err))
    },
  })

  const closeConn = useMutation({
    mutationFn: () => connectionApi.close(item.id),
    onSuccess: () => {
      const { activeConnectionId, setSelectedTable } = useWorkspaceStore.getState()
      if (activeConnectionId === item.id) {
        setActiveConnectionId(null)
        setSelectedTable(null)
      }
      onRefresh()
    },
    onError: (err) => setError(formatError(t, err)),
  })

  const removeConn = useMutation({
    mutationFn: async () => {
      const msg =
        item.status === 'open'
          ? t('connection.confirmRemoveOpen', { name: item.name })
          : t('connection.confirmRemove', { name: item.name })
      if (!window.confirm(msg)) throw new Error('cancelled')
      return connectionApi.remove(item.id)
    },
    onSuccess: onRefresh,
    onError: (err) => {
      if ((err as Error).message === 'cancelled') return
      setError(formatError(t, err))
    },
  })

  return (
    <div style={{ paddingLeft: depth * 12, ...dragStyle }} className="mb-1" ref={setNodeRef}>
      <ContextMenu>
        <ContextMenuTrigger asChild>
          <div
            className="cursor-default rounded border border-border/40 p-2 hover:bg-muted/20"
            data-sqlite-drop-target={item.type === 'sqlite' ? '' : undefined}
            data-connection-id={item.id}
            data-connection-type={item.type}
            data-connection-open={isOpen ? 'true' : 'false'}
            {...listeners}
            {...attributes}
            onDoubleClick={() => {
              if (isOpen) setActiveConnectionId(item.id)
              else openConn.mutate()
            }}
            onClick={() => setActiveConnectionId(item.id)}
          >
            <div className="flex items-center gap-1 text-sm font-medium">
              <span className="rounded bg-muted/40 px-1 text-[10px] font-bold text-muted">
                {connectionTypeLabel(item.type)}
              </span>
              <span className={isOpen ? 'text-green-500' : 'text-muted'}>{isOpen ? '●' : '○'}</span>
              <span className="truncate">{item.name}</span>
            </div>
            <p className="truncate text-xs text-muted" title={subtitle}>
              {subtitle}
            </p>
            {error && <p className="text-xs text-red-500">{error}</p>}
          </div>
        </ContextMenuTrigger>
        <ContextMenuContent>
          {!isOpen ? (
            <ContextMenuItem onSelect={() => openConn.mutate()}>
              {t('connection.openConnection')}
            </ContextMenuItem>
          ) : (
            <ContextMenuItem onSelect={() => closeConn.mutate()}>
              {t('connection.closeConnection')}
            </ContextMenuItem>
          )}
          <ContextMenuItem disabled={isOpen} onSelect={() => !isOpen && setEditOpen(true)}>
            {t('connection.edit')}
          </ContextMenuItem>
          {item.type === 'sqlite' && (
            <ContextMenuItem disabled={!isOpen} onSelect={() => isOpen && setAttachOpen(true)}>
              {t('connection.attachDatabase')}
            </ContextMenuItem>
          )}
          <ContextMenuSeparator />
          <ContextMenuItem variant="danger" onSelect={() => removeConn.mutate()}>
            {t('connection.delete')}
          </ContextMenuItem>
        </ContextMenuContent>
      </ContextMenu>
      {isOpen && (
        <ConnectionSchemaTree
          item={item}
          onSelectTable={(table, ctx) => selectTable(item.id, table, ctx)}
        />
      )}
      <EditConnectionDialog
        item={item}
        open={editOpen}
        onOpenChange={setEditOpen}
        onSaved={onRefresh}
      />
      {item.type === 'sqlite' && (
        <AttachDatabaseDialog
          connectionId={item.id}
          open={attachOpen}
          onOpenChange={setAttachOpen}
          onAttached={() => {
            void qc.invalidateQueries({ queryKey: ['namespaces', item.id] })
            onRefresh()
          }}
        />
      )}
    </div>
  )
}
