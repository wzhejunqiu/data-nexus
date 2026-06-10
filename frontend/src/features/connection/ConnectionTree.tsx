import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
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
import type { ConnectionGroupNode as GroupNode, SidebarRootItemRef } from '@/lib/types'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import type { AttachDatabaseOpen } from './attachDatabaseState'
import { ConnectionGroupNode } from './ConnectionGroupNode'
import { ConnectionTreeItem } from './ConnectionTreeItem'
import { VaultDialog, type VaultDialogMode } from './VaultDialog'
import { SidebarDndContext, RootDropZone } from './dnd/SidebarDndContext'
import { collectGroupMemberMaps, defaultRootItems } from './dnd/sidebarOrder'
import { useSidebarFileDrop } from './dnd/useSidebarFileDrop'
import { groupMatchesSearch, matchesConnectionName } from './sidebarSearch'
import { invokeSidebarRename, isAnySidebarRenaming } from './sidebarRenameHandlers'

function isEditableTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) return false
  const tag = target.tagName
  if (tag === 'INPUT' || tag === 'TEXTAREA') return true
  if (target.isContentEditable) return true
  return Boolean(target.closest('.monaco-editor'))
}

function countGroupConnections(groups: GroupNode[]): number {
  let total = 0
  for (const g of groups) {
    total += g.connections?.length ?? 0
    total += countGroupConnections(g.childGroups ?? [])
  }
  return total
}

export function ConnectionTree({
  onOpenAttach = () => {},
}: {
  onOpenAttach?: (target: AttachDatabaseOpen) => void
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const pushToast = useToastStore((s) => s.push)
  const sidebarFocus = useWorkspaceStore((s) => s.sidebarFocus)
  const sidebarRef = useRef<HTMLDivElement>(null)
  const [sidebarWidth, setSidebarWidth] = useState(260)
  const [vaultOpen, setVaultOpen] = useState(false)
  const [vaultMode, setVaultMode] = useState<VaultDialogMode>('unlock')
  const [pendingOpenId, setPendingOpenId] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [pendingRenameGroupId, setPendingRenameGroupId] = useState<string | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['connectionSidebarTree'],
    queryFn: () => connectionGroupApi.getSidebarTree(),
  })

  const refresh = useCallback(() => {
    qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
    qc.invalidateQueries({ queryKey: ['connections'] })
  }, [qc])

  useSidebarFileDrop({
    onAttach: ({ connectionId, paths }) => {
      onOpenAttach({
        connectionId,
        initialPath: paths[0],
        queue: paths.slice(1),
      })
    },
    onRefresh: refresh,
  })

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key !== 'F2' && e.key !== 'Enter') return
      if (isAnySidebarRenaming()) return
      if (isEditableTarget(e.target)) return
      if (document.querySelector('[role="dialog"]')) return
      if (!sidebarFocus) return
      e.preventDefault()
      invokeSidebarRename(sidebarFocus)
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [sidebarFocus])

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

  const groups = useMemo(() => data?.groups ?? [], [data?.groups])
  const freeConnections = useMemo(() => data?.freeConnections ?? [], [data?.freeConnections])
  const rootItems = useMemo(() => (data ? defaultRootItems(data) : []), [data])
  const groupMembers = useMemo(() => collectGroupMemberMaps(groups), [groups])

  const totalConnections = freeConnections.length + countGroupConnections(groups)
  const hasNoConnections = !isLoading && totalConnections === 0

  const isEmpty =
    !isLoading &&
    groups.length <= 1 &&
    freeConnections.length === 0 &&
    groups.every((g) => !g.connections?.length && !g.childGroups?.length)

  const filteredRootItems = useMemo(() => {
    const q = searchQuery.trim()
    if (!q) return rootItems
    return rootItems.filter((ref) => {
      if (ref.itemType === 'connection') {
        const item = freeConnections.find((c) => c.id === ref.itemId)
        return item && matchesConnectionName(item, q)
      }
      const group = groups.find((g) => g.id === ref.itemId)
      return group && groupMatchesSearch(group, q)
    })
  }, [rootItems, searchQuery, freeConnections, groups])

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
          searchQuery={searchQuery}
          pendingRenameGroupId={pendingRenameGroupId}
          onPendingRenameHandled={() => setPendingRenameGroupId(null)}
          onRequestRename={setPendingRenameGroupId}
          onRefresh={refresh}
          onVaultRequired={handleVaultRequired}
          onOpenAttach={onOpenAttach}
        />
      )
    }
    const item = freeConnections.find((c) => c.id === ref.itemId)
    if (!item) return null
    if (searchQuery.trim() && !matchesConnectionName(item, searchQuery)) return null
    return (
      <ConnectionTreeItem
        key={item.id}
        item={item}
        depth={0}
        sortContainerId="root"
        searchQuery={searchQuery}
        onRefresh={refresh}
        onVaultRequired={handleVaultRequired}
        onOpenAttach={onOpenAttach}
      />
    )
  }

  return (
    <>
      <aside
        className="relative flex flex-col border-r border-border"
        style={{ width: sidebarWidth, minWidth: 200, maxWidth: 480 }}
      >
        <div
          ref={sidebarRef}
          tabIndex={-1}
          className="flex min-h-0 flex-1 flex-col overflow-auto p-2 outline-none"
        >
          <input
            type="search"
            className="mb-2 w-full rounded border border-border bg-transparent px-2 py-1 text-sm"
            placeholder={t('connection.search')}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          <SidebarDndContext onRefresh={refresh} rootItems={rootItems} groupMembers={groupMembers}>
            <ContextMenu>
              <ContextMenuTrigger asChild>
                <div className="min-h-full">
                  <RootDropZone>
                    {isLoading && <p className="text-sm text-muted">{t('common.loading')}</p>}
                    {groups.length === 0 && !isLoading && (
                      <p className="text-sm text-muted">{t('connectionGroup.noGroupsHint')}</p>
                    )}
                    {hasNoConnections && groups.length > 0 && (
                      <p className="text-sm text-muted">{t('connection.emptyHint')}</p>
                    )}
                    {isEmpty && groups.length > 0 && !hasNoConnections && (
                      <p className="text-sm text-muted">{t('connectionGroup.emptyHint')}</p>
                    )}
                    {filteredRootItems.map(renderRootItem)}
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
        <div className="border-t border-border">
          <SavedQueries />
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
    </>
  )
}
