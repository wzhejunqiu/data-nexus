import { useDroppable } from '@dnd-kit/core'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/ContextMenu'
import { connectionGroupApi } from '@/lib/api/connectionGroup'
import { formatError } from '@/lib/api/errors'
import { useToastStore } from '@/components/ui/Toast'
import type { ConnectionGroupNode as GroupNode } from '@/lib/types'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import type { AttachDatabaseOpen } from './attachDatabaseState'
import { DeleteGroupDialog } from './DeleteGroupDialog'
import { ConnectionTreeItem } from './ConnectionTreeItem'
import type { VaultDialogMode } from './VaultDialog'
import { GroupSortableList } from './dnd/SidebarDndContext'
import { defaultMemberItems } from './dnd/sidebarOrder'
import { groupMatchesSearch, matchesConnectionName } from './sidebarSearch'
import { registerGroupRenameHandler } from './sidebarRenameHandlers'

const DEFAULT_GROUP_NAME = 'My Connections'

export function ConnectionGroupNode({
  node,
  depth,
  sortContainerId,
  searchQuery = '',
  pendingRenameGroupId,
  onPendingRenameHandled,
  onRequestRename,
  onRefresh,
  onVaultRequired,
  onOpenAttach = () => {},
}: {
  node: GroupNode
  depth: number
  sortContainerId: string
  searchQuery?: string
  pendingRenameGroupId?: string | null
  onPendingRenameHandled?: () => void
  onRequestRename?: (id: string) => void
  onRefresh: () => void
  onVaultRequired?: (id: string, mode: VaultDialogMode) => void
  onOpenAttach?: (target: AttachDatabaseOpen) => void
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const qc = useQueryClient()
  const [editing, setEditing] = useState(false)
  const [editName, setEditName] = useState(node.name)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [expanded, setExpanded] = useState(true)
  const sidebarFocus = useWorkspaceStore((s) => s.sidebarFocus)
  const setSelectedGroupId = useWorkspaceStore((s) => s.setSelectedGroupId)
  const setSidebarFocus = useWorkspaceStore((s) => s.setSidebarFocus)
  const isFocused = sidebarFocus?.kind === 'group' && sidebarFocus.id === node.id
  const q = searchQuery.trim()
  const visible = !q || groupMatchesSearch(node, q)

  const displayName =
    node.name === DEFAULT_GROUP_NAME ? t('connectionGroup.defaultName') : node.name

  const memberItems = useMemo(() => defaultMemberItems(node), [node])
  const memberIds = memberItems.map((m) => m.memberId)

  const connById = useMemo(
    () => new Map(node.connections?.map((c) => [c.id, c]) ?? []),
    [node.connections],
  )
  const childById = useMemo(
    () => new Map(node.childGroups?.map((g) => [g.id, g]) ?? []),
    [node.childGroups],
  )

  const {
    attributes,
    listeners,
    setNodeRef: sortRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: node.id,
    data: { type: 'group', sortable: { containerId: sortContainerId } },
  })
  const { setNodeRef: dropRef, isOver } = useDroppable({ id: node.id, data: { type: 'group' } })
  const setRefs = (el: HTMLDivElement | null) => {
    sortRef(el)
    dropRef(el)
  }
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const pendingRename = pendingRenameGroupId === node.id
  const isRenaming = editing || pendingRename

  const finishPendingRename = useCallback(() => {
    if (pendingRename) onPendingRenameHandled?.()
  }, [pendingRename, onPendingRenameHandled])

  const renameMut = useMutation({
    mutationFn: (name: string) => connectionGroupApi.renameGroup(node.id, name),
    onSuccess: () => {
      setEditing(false)
      finishPendingRename()
      onRefresh()
      qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
    },
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

  const createChildMut = useMutation({
    mutationFn: () => connectionGroupApi.createGroup(node.id, t('connectionGroup.newGroupName')),
    onSuccess: (created) => {
      onRefresh()
      qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
      if (created?.id) {
        setSidebarFocus({ kind: 'group', id: created.id })
        onRequestRename?.(created.id)
      }
    },
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

  const commitRename = () => {
    const trimmed = editName.trim()
    if (!trimmed) {
      setEditName(node.name)
      setEditing(false)
      finishPendingRename()
      return
    }
    if (trimmed !== node.name) renameMut.mutate(trimmed)
    else {
      setEditing(false)
      finishPendingRename()
    }
  }

  const startRename = useCallback(() => {
    if (editing) return
    setEditName(node.name)
    setEditing(true)
  }, [editing, node.name])

  useEffect(() => {
    return registerGroupRenameHandler(node.id, startRename)
  }, [node.id, startRename])

  const selectGroup = () => {
    setSelectedGroupId(node.id)
    setSidebarFocus({ kind: 'group', id: node.id })
  }

  if (!visible) return null

  return (
    <div style={{ paddingLeft: depth * 12 }} className="mb-1">
      <div ref={setRefs} style={style}>
        <ContextMenu>
          <ContextMenuTrigger asChild>
            <div
              {...listeners}
              {...attributes}
              data-group-drop-target=""
              data-group-id={node.id}
              className={`flex cursor-grab items-center gap-1 rounded px-1 py-0.5 hover:bg-muted/30 ${
                isOver ? 'ring-1 ring-accent' : ''
              } ${isFocused ? 'bg-muted/20 ring-1 ring-accent' : ''}`}
              onClick={selectGroup}
            >
              <button
                type="button"
                className="text-xs text-muted"
                onClick={(e) => {
                  e.stopPropagation()
                  setExpanded((v) => !v)
                }}
                aria-label="toggle"
              >
                {expanded ? '▼' : '▶'}
              </button>
              <span className="text-sm">📁</span>
              {isRenaming ? (
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
                      setEditName(node.name)
                      setEditing(false)
                      finishPendingRename()
                    }
                  }}
                  onBlur={commitRename}
                />
              ) : (
                <span className="truncate text-sm font-medium">{displayName}</span>
              )}
            </div>
          </ContextMenuTrigger>
          <ContextMenuContent>
            <ContextMenuItem onSelect={() => createChildMut.mutate()}>
              {t('connectionGroup.new')}
            </ContextMenuItem>
            <ContextMenuItem onSelect={startRename}>{t('connectionGroup.rename')}</ContextMenuItem>
            <ContextMenuSeparator />
            <ContextMenuItem variant="danger" onSelect={() => setDeleteOpen(true)}>
              {t('connectionGroup.delete')}
            </ContextMenuItem>
          </ContextMenuContent>
        </ContextMenu>
      </div>
      {expanded && (
        <div className="ml-2">
          <GroupSortableList groupId={node.id} memberIds={memberIds}>
            {memberItems.map((member) => {
              if (member.memberType === 'connection') {
                const conn = connById.get(member.memberId)
                if (!conn) return null
                if (q && !matchesConnectionName(conn, q)) return null
                return (
                  <ConnectionTreeItem
                    key={conn.id}
                    item={conn}
                    depth={depth + 1}
                    sortContainerId={node.id}
                    searchQuery={searchQuery}
                    onRefresh={onRefresh}
                    onVaultRequired={onVaultRequired}
                    onOpenAttach={onOpenAttach}
                  />
                )
              }
              const child = childById.get(member.memberId)
              if (!child) return null
              return (
                <ConnectionGroupNode
                  key={child.id}
                  node={child}
                  depth={depth + 1}
                  sortContainerId={node.id}
                  searchQuery={searchQuery}
                  pendingRenameGroupId={pendingRenameGroupId}
                  onPendingRenameHandled={onPendingRenameHandled}
                  onRequestRename={onRequestRename}
                  onRefresh={onRefresh}
                  onVaultRequired={onVaultRequired}
                  onOpenAttach={onOpenAttach}
                />
              )
            })}
          </GroupSortableList>
        </div>
      )}
      <DeleteGroupDialog
        groupId={node.id}
        groupName={displayName}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        onDeleted={onRefresh}
      />
    </div>
  )
}
