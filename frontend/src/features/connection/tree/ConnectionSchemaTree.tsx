import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
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
import { schemaApi } from '@/lib/api/schema'
import { isSystemDatabase } from '@/lib/systemDatabases'
import { tableKey } from '@/lib/tableKey'
import type { ConnectionListItem, DriverType, NamespaceInfo, TableInfo } from '@/lib/types'
import { useWorkspaceStore } from '@/stores/workspaceStore'

function TreeChevron({
  open,
  onToggle,
}: {
  open: boolean
  onToggle: (e: React.MouseEvent) => void
}) {
  return (
    <button
      type="button"
      className="shrink-0 px-0.5 text-muted hover:text-foreground"
      aria-label={open ? 'collapse' : 'expand'}
      onClick={onToggle}
    >
      {open ? '▼' : '▶'}
    </button>
  )
}

export function ConnectionSchemaTree({
  item,
  onSelectTable,
  searchQuery = '',
}: {
  item: ConnectionListItem
  onSelectTable: (table: string, ctx?: { database?: string; schema?: string }) => void
  searchQuery?: string
}) {
  const { t } = useTranslation()
  const [expandedNs, setExpandedNs] = useState<Record<string, boolean>>({})
  const [expandedSchema, setExpandedSchema] = useState<Record<string, boolean>>({})

  const { data: namespaces } = useQuery({
    queryKey: ['namespaces', item.id],
    queryFn: () => schemaApi.listNamespaces(item.id),
    enabled: item.status === 'open',
  })

  const nsItems = namespaces?.items ?? []
  const q = searchQuery.trim().toLowerCase()

  return (
    <div className="ml-3 mt-1 border-l border-border/50 pl-2">
      {nsItems.map((ns) => {
        const nsKey = ns.name
        const autoExpand = Boolean(q)
        const nsOpen = autoExpand || (expandedNs[nsKey] ?? false)
        return (
          <div key={nsKey} className="mb-1">
            <NamespaceRow
              ns={ns}
              nsOpen={nsOpen}
              connectionId={item.id}
              driverType={item.type}
              onToggleExpand={() => setExpandedNs((s) => ({ ...s, [nsKey]: !nsOpen }))}
              onExpand={() => setExpandedNs((s) => ({ ...s, [nsKey]: true }))}
            />
            {nsOpen && item.type === 'postgres' && (
              <PostgresSchemas
                connectionId={item.id}
                database={ns.name}
                expandedSchema={expandedSchema}
                setExpandedSchema={setExpandedSchema}
                onSelectTable={onSelectTable}
                searchQuery={q}
              />
            )}
            {nsOpen && item.type !== 'postgres' && (
              <NamespaceTables
                connectionId={item.id}
                database={ns.name}
                onSelectTable={onSelectTable}
                searchQuery={q}
              />
            )}
          </div>
        )
      })}
      {nsItems.length === 0 && <p className="text-xs text-muted">{t('common.loading')}</p>}
    </div>
  )
}

function NamespaceRow({
  ns,
  nsOpen,
  connectionId,
  driverType,
  onToggleExpand,
  onExpand,
}: {
  ns: NamespaceInfo
  nsOpen: boolean
  connectionId: string
  driverType: DriverType
  onToggleExpand: () => void
  onExpand: () => void
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

  if (ns.kind !== 'attach') return row

  return (
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
  )
}

function PostgresSchemas({
  connectionId,
  database,
  expandedSchema,
  setExpandedSchema,
  onSelectTable,
  searchQuery,
}: {
  connectionId: string
  database: string
  expandedSchema: Record<string, boolean>
  setExpandedSchema: React.Dispatch<React.SetStateAction<Record<string, boolean>>>
  onSelectTable: (table: string, ctx?: { database?: string; schema?: string }) => void
  searchQuery: string
}) {
  const setBrowseContext = useWorkspaceStore((s) => s.setBrowseContext)
  const browseDatabase = useWorkspaceStore((s) => s.browseContext[connectionId]?.database)
  const browseSchema = useWorkspaceStore((s) => s.browseContext[connectionId]?.schema)

  const { data } = useQuery({
    queryKey: ['schemas', connectionId, database],
    queryFn: () => schemaApi.listSchemas(connectionId, database),
  })
  const schemas = data?.items ?? []
  const q = searchQuery.trim().toLowerCase()

  return (
    <div className="ml-3">
      {schemas.map((sch) => {
        const key = `${database}/${sch.name}`
        const autoExpand = Boolean(q)
        const open = autoExpand || (expandedSchema[key] ?? false)
        const isActiveBrowse = browseDatabase === database && browseSchema === sch.name

        const setBrowse = () => setBrowseContext(connectionId, { database, schema: sch.name })

        return (
          <div key={key}>
            <div
              className={`flex w-full items-center gap-1 text-left text-xs text-muted hover:text-foreground ${
                isActiveBrowse ? 'rounded ring-1 ring-accent/50' : ''
              }`}
              onClick={() => setBrowse()}
              onDoubleClick={() => {
                setBrowse()
                setExpandedSchema((s) => ({ ...s, [key]: true }))
              }}
            >
              <TreeChevron
                open={open}
                onToggle={(e) => {
                  e.stopPropagation()
                  setExpandedSchema((s) => ({ ...s, [key]: !open }))
                }}
              />
              <span>{sch.name}</span>
            </div>
            {open && (
              <NamespaceTables
                connectionId={connectionId}
                database={database}
                schema={sch.name}
                onSelectTable={onSelectTable}
                searchQuery={q}
              />
            )}
          </div>
        )
      })}
    </div>
  )
}

function NamespaceTables({
  connectionId,
  database,
  schema,
  onSelectTable,
  searchQuery = '',
}: {
  connectionId: string
  database: string
  schema?: string
  onSelectTable: (table: string, ctx?: { database?: string; schema?: string }) => void
  searchQuery?: string
}) {
  const { t } = useTranslation()
  const { data } = useQuery({
    queryKey: ['tables', connectionId, database, schema],
    queryFn: () => schemaApi.listTables(connectionId, { database, schema }),
  })
  const tables = data?.items ?? []
  const q = searchQuery.trim().toLowerCase()

  const filtered = q ? tables.filter((tbl) => tbl.name.toLowerCase().includes(q)) : tables

  if (filtered.length === 0)
    return <p className="ml-4 text-xs text-muted">{t('connection.emptyTables')}</p>

  const tableItems = filtered.filter((tbl) => tbl.type !== 'view')
  const viewItems = filtered.filter((tbl) => tbl.type === 'view')

  return (
    <div className="ml-4">
      {tableItems.length > 0 && (
        <TableGroup
          label={`${t('connection.tables')} (${tableItems.length})`}
          items={tableItems}
          database={database}
          schema={schema}
          onSelectTable={onSelectTable}
        />
      )}
      {viewItems.length > 0 && (
        <TableGroup
          label={`${t('connection.views')} (${viewItems.length})`}
          items={viewItems}
          database={database}
          schema={schema}
          onSelectTable={onSelectTable}
        />
      )}
    </div>
  )
}

function TableGroup({
  label,
  items,
  database,
  schema,
  onSelectTable,
}: {
  label: string
  items: TableInfo[]
  database: string
  schema?: string
  onSelectTable: (table: string, ctx?: { database?: string; schema?: string }) => void
}) {
  return (
    <div className="mb-1">
      <p className="text-[10px] font-semibold uppercase tracking-wide text-muted">{label}</p>
      <ul>
        {items.map((tbl) => (
          <li key={tbl.name}>
            <button
              type="button"
              className="w-full truncate text-left text-xs hover:text-accent"
              onClick={() => onSelectTable(tableKey(tbl), { database, schema })}
            >
              {tbl.type === 'view' ? '👁' : '📋'} {tbl.name}
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
