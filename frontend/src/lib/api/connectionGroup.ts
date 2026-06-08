import {
  CountConnectionsInGroup,
  CreateGroup,
  DeleteGroup,
  GetSidebarTree,
  MoveConnectionToGroup,
  MoveGroup,
  ReleaseConnection,
  RenameGroup,
  ReorderGroupMembers,
  ReorderSidebarRoot,
} from '../../../wailsjs/go/wails/ConnectionGroupService'
import type {
  ConnectionSidebarTree,
  DeleteGroupRequest,
  GroupMemberRef,
  MoveGroupRequest,
  SidebarRootItemRef,
} from '../types'
import { mapWailsError } from './errors'

async function wrap<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    throw mapWailsError(err)
  }
}

export const connectionGroupApi = {
  getSidebarTree: () => wrap(() => GetSidebarTree() as Promise<ConnectionSidebarTree>),
  createGroup: (parentId: string, name: string) => wrap(() => CreateGroup(parentId, name)),
  renameGroup: (id: string, name: string) => wrap(() => RenameGroup(id, name)),
  countConnectionsInGroup: (id: string) => wrap(() => CountConnectionsInGroup(id)),
  deleteGroup: (req: DeleteGroupRequest) => wrap(() => DeleteGroup(req)),
  moveGroup: (req: MoveGroupRequest) =>
    wrap(() =>
      MoveGroup({
        id: req.id,
        newParentId: req.newParentId ?? undefined,
        sortOrder: req.sortOrder,
      }),
    ),
  moveConnectionToGroup: (connectionId: string, groupId: string, sortOrder: number) =>
    wrap(() => MoveConnectionToGroup(connectionId, groupId, sortOrder)),
  releaseConnection: (connectionId: string, sortOrder: number) =>
    wrap(() => ReleaseConnection(connectionId, sortOrder)),
  reorderGroupMembers: (groupId: string, ordered: GroupMemberRef[]) =>
    wrap(() => ReorderGroupMembers(groupId, ordered)),
  reorderSidebarRoot: (ordered: SidebarRootItemRef[]) => wrap(() => ReorderSidebarRoot(ordered)),
}
