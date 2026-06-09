type RenameHandler = () => void

const groupHandlers = new Map<string, RenameHandler>()
const connectionHandlers = new Map<string, RenameHandler>()

export function registerGroupRenameHandler(id: string, handler: RenameHandler) {
  groupHandlers.set(id, handler)
  return () => {
    groupHandlers.delete(id)
  }
}

export function registerConnectionRenameHandler(id: string, handler: RenameHandler) {
  connectionHandlers.set(id, handler)
  return () => {
    connectionHandlers.delete(id)
  }
}

export function invokeSidebarRename(focus: { kind: 'group' | 'connection'; id: string }) {
  if (focus.kind === 'group') {
    groupHandlers.get(focus.id)?.()
  } else {
    connectionHandlers.get(focus.id)?.()
  }
}

export function isAnySidebarRenaming() {
  return document.querySelector('[data-sidebar-rename-input="true"]') !== null
}
