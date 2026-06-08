import { arrayMove } from '@dnd-kit/sortable'
import type {
  ConnectionGroupNode,
  ConnectionSidebarTree,
  GroupMemberRef,
  SidebarRootItemRef,
} from '@/lib/types'

export function defaultRootItems(tree: ConnectionSidebarTree): SidebarRootItemRef[] {
  if (tree.rootItems?.length) return tree.rootItems
  return [
    ...tree.groups.map((g) => ({ itemType: 'group' as const, itemId: g.id })),
    ...tree.freeConnections.map((c) => ({ itemType: 'connection' as const, itemId: c.id })),
  ]
}

export function defaultMemberItems(node: ConnectionGroupNode): GroupMemberRef[] {
  if (node.memberItems?.length) return node.memberItems
  return [
    ...(node.connections?.map((c) => ({ memberType: 'connection' as const, memberId: c.id })) ??
      []),
    ...(node.childGroups?.map((g) => ({ memberType: 'group' as const, memberId: g.id })) ?? []),
  ]
}

export function collectGroupMemberMaps(
  groups: ConnectionGroupNode[],
): Record<string, GroupMemberRef[]> {
  const map: Record<string, GroupMemberRef[]> = {}
  const walk = (g: ConnectionGroupNode) => {
    map[g.id] = defaultMemberItems(g)
    g.childGroups?.forEach(walk)
  }
  groups.forEach(walk)
  return map
}

export function reorderRootItems(
  items: SidebarRootItemRef[],
  activeId: string,
  overId: string,
): SidebarRootItemRef[] {
  const oldIndex = items.findIndex((i) => i.itemId === activeId)
  const newIndex = items.findIndex((i) => i.itemId === overId)
  if (oldIndex < 0 || newIndex < 0 || oldIndex === newIndex) return items
  return arrayMove(items, oldIndex, newIndex)
}

export function reorderMemberItems(
  items: GroupMemberRef[],
  activeId: string,
  overId: string,
): GroupMemberRef[] {
  const oldIndex = items.findIndex((i) => i.memberId === activeId)
  const newIndex = items.findIndex((i) => i.memberId === overId)
  if (oldIndex < 0 || newIndex < 0 || oldIndex === newIndex) return items
  return arrayMove(items, oldIndex, newIndex)
}
