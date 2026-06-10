import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useState } from 'react'
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
import type { AttachDatabaseOpen } from './attachDatabaseState'
import { EditConnectionDialog } from './EditConnectionDialog'
import type { VaultDialogMode } from './VaultDialog'
import { ConnectionSchemaTree } from './tree/ConnectionSchemaTree'
import { registerConnectionRenameHandler } from './sidebarRenameHandlers'

export function ConnectionTreeItem({
  item,
  depth,
  sortContainerId,
  searchQuery = '',
  onRefresh,
  onVaultRequired,
  onOpenAttach = () => {},
}: {
  item: ConnectionListItem
  depth: number
  sortContainerId: string
  searchQuery?: string
  onRefresh: () => void
  onVaultRequired?: (id: string, mode: VaultDialogMode) => void
  onOpenAttach?: (target: AttachDatabaseOpen) => void
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const isOpen = item.status === 'open'
  const subtitle = connectionSubtitle(item)
  const setActiveConnectionId = useWorkspaceStore((s) => s.setActiveConnectionId)
  const setSidebarFocus = useWorkspaceStore((s) => s.setSidebarFocus)
  const sidebarFocus = useWorkspaceStore((s) => s.sidebarFocus)
  const selectTable = useWorkspaceStore((s) => s.selectTable)
  const isFocused = sidebarFocus?.kind === 'connection' && sidebarFocus.id === item.id
  const [editOpen, setEditOpen] = useState(false)
  const [editing, setEditing] = useState(false)
  const [editName, setEditName] = useState(item.name)
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

  const renameMut = useMutation({
    mutationFn: (name: string) => connectionApi.rename(item.id, name),
    onSuccess: () => {
      setEditing(false)
      onRefresh()
      qc.invalidateQueries({ queryKey: ['connections'] })
      qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
    },
    onError: (err) => setError(formatError(t, err)),
  })

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

  const commitRename = () => {
    const trimmed = editName.trim()
    if (!trimmed) {
      setEditName(item.name)
      setEditing(false)
      return
    }
    if (trimmed !== item.name) renameMut.mutate(trimmed)
    else setEditing(false)
  }

  const startRename = useCallback(() => {
    if (editing) return
    setEditName(item.name)
    setEditing(true)
  }, [editing, item.name])

  useEffect(() => {
    return registerConnectionRenameHandler(item.id, startRename)
  }, [item.id, startRename])

  const selectConnection = () => {
    setActiveConnectionId(item.id)
    setSidebarFocus({ kind: 'connection', id: item.id })
  }

  return (
    <div style={{ paddingLeft: depth * 12, ...dragStyle }} className="mb-1" ref={setNodeRef}>
      <ContextMenu>
        <ContextMenuTrigger asChild>
          <div
            className={`cursor-default rounded border border-border/40 p-2 hover:bg-muted/20 ${
              isFocused ? 'bg-muted/20 ring-1 ring-accent' : ''
            }`}
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
            onClick={selectConnection}
          >
            <div className="flex items-center gap-1 text-sm font-medium">
              <span className="rounded bg-muted/40 px-1 text-[10px] font-bold text-muted">
                {connectionTypeLabel(item.type)}
              </span>
              <span className={isOpen ? 'text-green-500' : 'text-muted'}>{isOpen ? '●' : '○'}</span>
              {editing ? (
                <input
                  data-sidebar-rename-input="true"
                  className="min-w-0 flex-1 rounded border border-border bg-transparent px-1 text-sm"
                  value={editName}
                  autoFocus
                  onChange={(e) => setEditName(e.target.value)}
                  onClick={(e) => e.stopPropagation()}
                  onKeyDown={(e) => {
                    e.stopPropagation()
                    if (e.key === 'Enter') commitRename()
                    if (e.key === 'Escape') {
                      setEditName(item.name)
                      setEditing(false)
                    }
                  }}
                  onBlur={commitRename}
                />
              ) : (
                <span className="truncate">{item.name}</span>
              )}
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
          <ContextMenuItem onSelect={startRename}>{t('connection.rename')}</ContextMenuItem>
          <ContextMenuItem
            disabled={isOpen}
            hint={isOpen ? t('connection.editRequiresClosed') : undefined}
            onSelect={() => !isOpen && setEditOpen(true)}
          >
            {t('connection.edit')}
          </ContextMenuItem>
          {item.type === 'sqlite' && (
            <ContextMenuItem
              disabled={!isOpen}
              hint={!isOpen ? t('connection.attachRequiresOpen') : undefined}
              onSelect={() => isOpen && onOpenAttach({ connectionId: item.id })}
            >
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
          searchQuery={searchQuery}
          onSelectTable={(table, ctx) => selectTable(item.id, table, ctx)}
        />
      )}
      <EditConnectionDialog
        item={item}
        open={editOpen}
        onOpenChange={setEditOpen}
        onSaved={onRefresh}
      />
    </div>
  )
}
