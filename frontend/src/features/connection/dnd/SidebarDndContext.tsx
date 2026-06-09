import {
  DndContext,
  DragOverlay,
  PointerSensor,
  closestCenter,
  useDroppable,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { useToastStore } from '@/components/ui/Toast'
import { connectionGroupApi } from '@/lib/api/connectionGroup'
import { formatError } from '@/lib/api/errors'
import type { GroupMemberRef, SidebarRootItemRef } from '@/lib/types'
import { reorderMemberItems, reorderRootItems } from './sidebarOrder'

export type SidebarDragItem = {
  type: 'group' | 'connection'
  id: string
}

export function SidebarDndContext({
  children,
  onRefresh,
  rootItems,
  groupMembers,
}: {
  children: ReactNode
  onRefresh: () => void
  rootItems: SidebarRootItemRef[]
  groupMembers: Record<string, GroupMemberRef[]>
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const [active, setActive] = useState<SidebarDragItem | null>(null)
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }))

  const onDragEnd = async (event: DragEndEvent) => {
    setActive(null)
    const { active: dragActive, over } = event
    if (!over || dragActive.id === over.id) return

    const fromType = String(dragActive.data.current?.type ?? '')
    const fromId = String(dragActive.id)
    const overType = String(over.data.current?.type ?? '')
    const overId = String(over.id)
    const activeContainer = dragActive.data.current?.sortable?.containerId as string | undefined
    const overContainer = over.data.current?.sortable?.containerId as string | undefined

    try {
      if (activeContainer && overContainer && activeContainer === overContainer) {
        if (activeContainer === 'root') {
          const next = reorderRootItems(rootItems, fromId, overId)
          if (next !== rootItems) {
            await connectionGroupApi.reorderSidebarRoot(next)
            onRefresh()
          }
        } else {
          const members = groupMembers[activeContainer] ?? []
          const next = reorderMemberItems(members, fromId, overId)
          if (next !== members) {
            await connectionGroupApi.reorderGroupMembers(activeContainer, next)
            onRefresh()
          }
        }
        return
      }

      if (overType === 'connection' && fromType === 'connection' && overContainer) {
        if (overContainer === 'root') {
          await connectionGroupApi.releaseConnection(fromId, -1)
        } else {
          await connectionGroupApi.moveConnectionToGroup(fromId, overContainer, -1)
        }
      } else if (overType === 'group' && fromType === 'connection') {
        await connectionGroupApi.moveConnectionToGroup(fromId, overId, -1)
      } else if (overType === 'root' && fromType === 'connection') {
        await connectionGroupApi.releaseConnection(fromId, -1)
      } else if (overType === 'group' && fromType === 'group') {
        await connectionGroupApi.moveGroup({ id: fromId, newParentId: overId, sortOrder: -1 })
      } else if (overType === 'root' && fromType === 'group') {
        await connectionGroupApi.moveGroup({ id: fromId, newParentId: null, sortOrder: -1 })
      } else {
        return
      }
      onRefresh()
    } catch (err) {
      pushToast(formatError(t, err), 'error')
    }
  }

  const rootSortableIds = rootItems.map((i) => i.itemId)

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={(e) => {
        const type = e.active.data.current?.type
        if (type === 'group' || type === 'connection') {
          setActive({ type, id: String(e.active.id) })
        }
      }}
      onDragEnd={(e) => void onDragEnd(e)}
      onDragCancel={() => setActive(null)}
    >
      <SortableContext items={rootSortableIds} strategy={verticalListSortingStrategy}>
        {children}
      </SortableContext>
      <DragOverlay>
        {active ? (
          <div className="rounded border border-border bg-background px-2 py-1 text-xs shadow">
            {active.type === 'group' ? '📁' : '●'} {t('connectionGroup.dragging')}
          </div>
        ) : null}
      </DragOverlay>
    </DndContext>
  )
}

export function RootDropZone({ children }: { children: ReactNode }) {
  const { setNodeRef, isOver } = useDroppable({ id: 'sidebar-root', data: { type: 'root' } })
  return (
    <div ref={setNodeRef} className={isOver ? 'rounded bg-muted/20' : undefined}>
      {children}
    </div>
  )
}

export function GroupSortableList({
  groupId,
  memberIds,
  children,
}: {
  groupId: string
  memberIds: string[]
  children: ReactNode
}) {
  return (
    <SortableContext items={memberIds} strategy={verticalListSortingStrategy}>
      <div data-sortable-container={groupId}>{children}</div>
    </SortableContext>
  )
}
