import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '@/components/ui/ContextMenu'
import { useToastStore } from '@/components/ui/Toast'
import { appApi } from '@/lib/api/app'
import { connectionApi } from '@/lib/api/connection'
import { formatError } from '@/lib/api/errors'
import { isSystemDatabase } from '@/lib/systemDatabases'
import type { DriverType, NamespaceInfo } from '@/lib/types'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import { SchemaList } from './SchemaNode'
import { TreeChevron } from './TreeNodeRow'
import { NamespaceTables } from './namespaceTables'

export function NamespaceNode({
  ns,
  nsOpen,
  connectionId,
  driverType,
  searchQuery,
  expandedSchema,
  setExpandedSchema,
  onToggleExpand,
  onExpand,
  onSelectTable,
}: {
  ns: NamespaceInfo
  nsOpen: boolean
  connectionId: string
  driverType: DriverType
  searchQuery: string
  expandedSchema: Record<string, boolean>
  setExpandedSchema: React.Dispatch<React.SetStateAction<Record<string, boolean>>>
  onToggleExpand: () => void
  onExpand: () => void
  onSelectTable: (table: string, ctx?: { database?: string; schema?: string }) => void
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const pushToast = useToastStore((s) => s.push)
  const browseDatabase = useWorkspaceStore((s) => s.browseContext[connectionId]?.database)
  const browseSchema = useWorkspaceStore((s) => s.browseContext[connectionId]?.schema)
  const setBrowseContext = useWorkspaceStore((s) => s.setBrowseContext)
  const activeConnectionId = useWorkspaceStore((s) => s.activeConnectionId)
  const selectedTable = useWorkspaceStore((s) => s.selectedTable)
  const setSelectedTable = useWorkspaceStore((s) => s.setSelectedTable)

  const isActiveBrowse = browseDatabase === ns.name && !browseSchema
  const isSystem = isSystemDatabase(ns.name, driverType)

  const { data: platform } = useQuery({
    queryKey: ['platform'],
    queryFn: () => appApi.getPlatform(),
    staleTime: Infinity,
    enabled: ns.kind === 'attach',
  })

  const revealLabel =
    platform?.startsWith('darwin') === true
      ? t('attach.showInFinder')
      : t('attach.showInFileManager')

  const setBrowse = () => setBrowseContext(connectionId, { database: ns.name })

  const needsDetachConfirm = () => {
    if (browseDatabase === ns.name) return true
    if (activeConnectionId === connectionId && selectedTable?.startsWith(`${ns.name}.`)) {
      return true
    }
    return false
  }

  const detach = async () => {
    if (needsDetachConfirm() && !window.confirm(t('attach.detachConfirm'))) return
    try {
      await connectionApi.detach(connectionId, ns.name)
      if (needsDetachConfirm()) {
        setBrowseContext(connectionId, {})
        if (activeConnectionId === connectionId && selectedTable?.startsWith(`${ns.name}.`)) {
          setSelectedTable(null)
        }
      }
      await qc.invalidateQueries({ queryKey: ['namespaces', connectionId] })
      await qc.invalidateQueries({ queryKey: ['tables', connectionId] })
    } catch (err) {
      pushToast(formatError(t, err), 'error')
    }
  }

  const reveal = async () => {
    if (!ns.filePath) return
    try {
      await appApi.revealFileInExplorer(ns.filePath)
    } catch (err) {
      pushToast(formatError(t, err), 'error')
    }
  }

  const rowClass = [
    'flex w-full items-center gap-1 text-left text-xs hover:text-foreground',
    isSystem ? 'text-muted italic' : 'text-muted',
    isActiveBrowse ? 'rounded ring-1 ring-accent/50' : '',
  ].join(' ')

  const row = (
    <div
      className={rowClass}
      title={ns.kind === 'attach' && ns.filePath ? ns.filePath : undefined}
      onClick={() => setBrowse()}
      onDoubleClick={() => {
        setBrowse()
        onExpand()
      }}
    >
      <TreeChevron
        open={nsOpen}
        onToggle={(e) => {
          e.stopPropagation()
          onToggleExpand()
        }}
      />
      <span>{ns.kind === 'attach' ? '📎' : '🗄'}</span>
      <span className="truncate">{ns.name}</span>
      {ns.kind === 'attach' && (
        <span className="ml-1 rounded bg-muted/40 px-1 text-[10px]">
          {t('connection.attachedBadge')}
        </span>
      )}
    </div>
  )

  return (
    <div className="mb-1">
      {ns.kind === 'attach' ? (
        <ContextMenu>
          <ContextMenuTrigger asChild>{row}</ContextMenuTrigger>
          <ContextMenuContent>
            {ns.filePath && (
              <ContextMenuItem onSelect={() => void reveal()}>{revealLabel}</ContextMenuItem>
            )}
            <ContextMenuItem variant="danger" onSelect={() => void detach()}>
              {t('attach.detach')}
            </ContextMenuItem>
          </ContextMenuContent>
        </ContextMenu>
      ) : (
        row
      )}
      {nsOpen && driverType === 'postgres' && (
        <SchemaList
          connectionId={connectionId}
          database={ns.name}
          expandedSchema={expandedSchema}
          setExpandedSchema={setExpandedSchema}
          onSelectTable={onSelectTable}
          searchQuery={searchQuery}
        />
      )}
      {nsOpen && driverType !== 'postgres' && (
        <NamespaceTables
          connectionId={connectionId}
          database={ns.name}
          onSelectTable={onSelectTable}
          searchQuery={searchQuery}
        />
      )}
    </div>
  )
}
