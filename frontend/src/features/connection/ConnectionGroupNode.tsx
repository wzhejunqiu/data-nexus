import { useDroppable } from '@dnd-kit/core'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
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
import { DeleteGroupDialog } from './DeleteGroupDialog'
import { ConnectionTreeItem } from './ConnectionTreeItem'
import type { VaultDialogMode } from './VaultDialog'
import { GroupSortableList } from './dnd/SidebarDndContext'
import { defaultMemberItems } from './dnd/sidebarOrder'

const DEFAULT_GROUP_NAME = 'My Connections'

export function ConnectionGroupNode({
  node,
  depth,
  sortContainerId,
  onRefresh,
  onVaultRequired,
}: {
  node: GroupNode
  depth: number
  sortContainerId: string
  onRefresh: () => void
  onVaultRequired?: (id: string, mode: VaultDialogMode) => void
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const qc = useQueryClient()
  const [editing, setEditing] = useState(false)
  const [editName, setEditName] = useState(node.name)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [expanded, setExpanded] = useState(true)

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

  const renameMut = useMutation({
    mutationFn: (name: string) => connectionGroupApi.renameGroup(node.id, name),
    onSuccess: () => {
      setEditing(false)
      onRefresh()
      qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
    },
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

  const createChildMut = useMutation({
    mutationFn: () => connectionGroupApi.createGroup(node.id, t('connectionGroup.newGroupName')),
    onSuccess: () => {
      onRefresh()
      qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
    },
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

  const commitRename = () => {
    const trimmed = editName.trim()
    if (!trimmed) {
      setEditName(node.name)
      setEditing(false)
      return
    }
    if (trimmed !== node.name) renameMut.mutate(trimmed)
    else setEditing(false)
  }

  return (
    <div style={{ paddingLeft: depth * 12 }} className="mb-1">
      <div ref={setRefs} style={style}>
        <ContextMenu>
          <ContextMenuTrigger asChild>
            <div
              {...listeners}
              {...attributes}
              className={`flex cursor-grab items-center gap-1 rounded px-1 py-0.5 hover:bg-muted/30 ${isOver ? 'ring-1 ring-accent' : ''}`}
            >
              <button
                type="button"
                className="text-xs text-muted"
                onClick={() => setExpanded((e) => !e)}
                aria-label="toggle"
              >
                {expanded ? '▼' : '▶'}
              </button>
              <span className="text-sm">📁</span>
              {editing ? (
                <input
                  className="min-w-0 flex-1 rounded border border-border bg-transparent px-1 text-sm"
                  value={editName}
                  autoFocus
                  onChange={(e) => setEditName(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') commitRename()
                    if (e.key === 'Escape') {
                      setEditName(node.name)
                      setEditing(false)
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
            <ContextMenuItem
              onSelect={() => {
                setEditName(node.name)
                setEditing(true)
              }}
            >
              {t('connectionGroup.rename')}
            </ContextMenuItem>
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
                return (
                  <ConnectionTreeItem
                    key={conn.id}
                    item={conn}
                    depth={depth + 1}
                    sortContainerId={node.id}
                    onRefresh={onRefresh}
                    onVaultRequired={onVaultRequired}
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
                  onRefresh={onRefresh}
                  onVaultRequired={onVaultRequired}
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
