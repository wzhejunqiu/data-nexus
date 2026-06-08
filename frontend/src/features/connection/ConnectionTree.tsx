import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { useCallback, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '@/components/ui/ContextMenu'
import { useToastStore } from '@/components/ui/Toast'
import { connectionGroupApi } from '@/lib/api/connectionGroup'
import { formatError } from '@/lib/api/errors'
import { SavedQueries } from '@/features/saved-queries/SavedQueries'
import type { SidebarRootItemRef } from '@/lib/types'
import { ConnectionGroupNode } from './ConnectionGroupNode'
import { ConnectionTreeItem } from './ConnectionTreeItem'
import { AttachDatabaseDialog } from './AttachDatabaseDialog'
import { VaultDialog, type VaultDialogMode } from './VaultDialog'
import { SidebarDndContext, RootDropZone } from './dnd/SidebarDndContext'
import { collectGroupMemberMaps, defaultRootItems } from './dnd/sidebarOrder'
import { useSidebarFileDrop, type FileAttachTarget } from './dnd/useSidebarFileDrop'

export function ConnectionTree() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const pushToast = useToastStore((s) => s.push)
  const [sidebarWidth, setSidebarWidth] = useState(260)
  const [vaultOpen, setVaultOpen] = useState(false)
  const [vaultMode, setVaultMode] = useState<VaultDialogMode>('unlock')
  const [pendingOpenId, setPendingOpenId] = useState<string | null>(null)
  const [fileAttach, setFileAttach] = useState<FileAttachTarget | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['connectionSidebarTree'],
    queryFn: () => connectionGroupApi.getSidebarTree(),
  })

  const refresh = useCallback(() => {
    qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
    qc.invalidateQueries({ queryKey: ['connections'] })
  }, [qc])

  useSidebarFileDrop({
    onAttach: setFileAttach,
    onRefresh: refresh,
  })

  const createRootGroupMut = useMutation({
    mutationFn: () => connectionGroupApi.createGroup('', t('connectionGroup.newGroupName')),
    onSuccess: () => refresh(),
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

  const handleVaultRequired = (id: string, mode: VaultDialogMode) => {
    setPendingOpenId(id)
    setVaultMode(mode)
    setVaultOpen(true)
  }

  const groups = data?.groups ?? []
  const freeConnections = data?.freeConnections ?? []
  const rootItems = useMemo(() => (data ? defaultRootItems(data) : []), [data])
  const groupMembers = useMemo(() => collectGroupMemberMaps(groups), [groups])

  const isEmpty =
    !isLoading &&
    groups.length <= 1 &&
    freeConnections.length === 0 &&
    groups.every((g) => !g.connections?.length && !g.childGroups?.length)

  const advanceFileAttach = () => {
    setFileAttach((prev) => {
      if (!prev) return null
      void qc.invalidateQueries({ queryKey: ['namespaces', prev.connectionId] })
      const remaining = prev.paths.slice(1)
      if (remaining.length > 0) {
        return { connectionId: prev.connectionId, paths: remaining }
      }
      refresh()
      return null
    })
  }

  const renderRootItem = (ref: SidebarRootItemRef) => {
    if (ref.itemType === 'group') {
      const rootGroup = groups.find((g) => g.id === ref.itemId)
      if (!rootGroup) return null
      return (
        <ConnectionGroupNode
          key={rootGroup.id}
          node={rootGroup}
          depth={0}
          sortContainerId="root"
          onRefresh={refresh}
          onVaultRequired={handleVaultRequired}
        />
      )
    }
    const item = freeConnections.find((c) => c.id === ref.itemId)
    if (!item) return null
    return (
      <ConnectionTreeItem
        key={item.id}
        item={item}
        depth={0}
        sortContainerId="root"
        onRefresh={refresh}
        onVaultRequired={handleVaultRequired}
      />
    )
  }

  return (
    <>
      <aside
        className="relative flex flex-col border-r border-border"
        style={{ width: sidebarWidth, minWidth: 200, maxWidth: 480 }}
      >
        <SavedQueries />
        <div className="flex min-h-0 flex-1 flex-col overflow-auto p-2">
          <SidebarDndContext onRefresh={refresh} rootItems={rootItems} groupMembers={groupMembers}>
            <ContextMenu>
              <ContextMenuTrigger asChild>
                <div className="min-h-full">
                  <RootDropZone>
                    {isLoading && <p className="text-sm text-muted">{t('common.loading')}</p>}
                    {groups.length === 0 && !isLoading && (
                      <p className="text-sm text-muted">{t('connectionGroup.noGroupsHint')}</p>
                    )}
                    {isEmpty && groups.length > 0 && (
                      <p className="text-sm text-muted">{t('connectionGroup.emptyHint')}</p>
                    )}
                    {rootItems.map(renderRootItem)}
                  </RootDropZone>
                </div>
              </ContextMenuTrigger>
              <ContextMenuContent>
                <ContextMenuItem onSelect={() => createRootGroupMut.mutate()}>
                  {t('connectionGroup.createRoot')}
                </ContextMenuItem>
              </ContextMenuContent>
            </ContextMenu>
          </SidebarDndContext>
        </div>
        <div
          className="absolute right-0 top-0 h-full w-1 cursor-col-resize hover:bg-accent/40"
          onMouseDown={(e) => {
            e.preventDefault()
            const startX = e.clientX
            const startW = sidebarWidth
            const onMove = (ev: MouseEvent) =>
              setSidebarWidth(Math.min(480, Math.max(200, startW + ev.clientX - startX)))
            const onUp = () => {
              window.removeEventListener('mousemove', onMove)
              window.removeEventListener('mouseup', onUp)
            }
            window.addEventListener('mousemove', onMove)
            window.addEventListener('mouseup', onUp)
          }}
        />
      </aside>
      <VaultDialog
        open={vaultOpen}
        mode={vaultMode}
        onOpenChange={setVaultOpen}
        onSuccess={() => {
          if (pendingOpenId) {
            refresh()
            setPendingOpenId(null)
          }
        }}
      />
      {fileAttach && (
        <AttachDatabaseDialog
          connectionId={fileAttach.connectionId}
          open
          closeOnSuccess={false}
          initialPath={fileAttach.paths[0]}
          onOpenChange={(open) => {
            if (!open) setFileAttach(null)
          }}
          onAttached={advanceFileAttach}
        />
      )}
    </>
  )
}
