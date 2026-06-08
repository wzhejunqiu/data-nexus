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
import type { ConnectionListItem, NamespaceInfo } from '@/lib/types'

export function ConnectionSchemaTree({
  item,
  onSelectTable,
}: {
  item: ConnectionListItem
  onSelectTable: (table: string, ctx?: { database?: string; schema?: string }) => void
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

  return (
    <div className="ml-3 mt-1 border-l border-border/50 pl-2">
      {nsItems.map((ns) => {
        const nsKey = ns.name
        const nsOpen = expandedNs[nsKey] ?? false
        return (
          <div key={nsKey} className="mb-1">
            <NamespaceRow
              ns={ns}
              nsOpen={nsOpen}
              connectionId={item.id}
              onToggle={() => setExpandedNs((s) => ({ ...s, [nsKey]: !nsOpen }))}
            />
            {nsOpen && item.type === 'postgres' && (
              <PostgresSchemas
                connectionId={item.id}
                database={ns.name}
                expandedSchema={expandedSchema}
                setExpandedSchema={setExpandedSchema}
                onSelectTable={onSelectTable}
              />
            )}
            {nsOpen && item.type !== 'postgres' && (
              <NamespaceTables
                connectionId={item.id}
                database={ns.name}
                onSelectTable={onSelectTable}
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
  onToggle,
}: {
  ns: NamespaceInfo
  nsOpen: boolean
  connectionId: string
  onToggle: () => void
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const pushToast = useToastStore((s) => s.push)

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

  const detach = async () => {
    try {
      await connectionApi.detach(connectionId, ns.name)
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

  const button = (
    <button
      type="button"
      className="flex w-full items-center gap-1 text-left text-xs text-muted hover:text-foreground"
      title={ns.kind === 'attach' && ns.filePath ? ns.filePath : undefined}
      onClick={onToggle}
    >
      <span>{nsOpen ? '▼' : '▶'}</span>
      <span>{ns.kind === 'attach' ? '📎' : '🗄'}</span>
      <span className="truncate">{ns.name}</span>
      {ns.kind === 'attach' && (
        <span className="ml-1 rounded bg-muted/40 px-1 text-[10px]">
          {t('connection.attachedBadge')}
        </span>
      )}
    </button>
  )

  if (ns.kind !== 'attach') return button

  return (
    <ContextMenu>
      <ContextMenuTrigger asChild>{button}</ContextMenuTrigger>
      <ContextMenuContent>
        {ns.filePath && (
          <ContextMenuItem onSelect={() => void reveal()}>{revealLabel}</ContextMenuItem>
        )}
        <ContextMenuItem onSelect={() => void detach()}>{t('attach.detach')}</ContextMenuItem>
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
}: {
  connectionId: string
  database: string
  expandedSchema: Record<string, boolean>
  setExpandedSchema: React.Dispatch<React.SetStateAction<Record<string, boolean>>>
  onSelectTable: (table: string, ctx?: { database?: string; schema?: string }) => void
}) {
  const { data } = useQuery({
    queryKey: ['schemas', connectionId, database],
    queryFn: () => schemaApi.listSchemas(connectionId, database),
  })
  const schemas = data?.items ?? []
  return (
    <div className="ml-3">
      {schemas.map((sch) => {
        const key = `${database}/${sch.name}`
        const open = expandedSchema[key] ?? false
        return (
          <div key={key}>
            <button
              type="button"
              className="flex w-full items-center gap-1 text-left text-xs text-muted hover:text-foreground"
              onClick={() => setExpandedSchema((s) => ({ ...s, [key]: !open }))}
            >
              <span>{open ? '▼' : '▶'}</span>
              <span>{sch.name}</span>
            </button>
            {open && (
              <NamespaceTables
                connectionId={connectionId}
                database={database}
                schema={sch.name}
                onSelectTable={onSelectTable}
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
}: {
  connectionId: string
  database: string
  schema?: string
  onSelectTable: (table: string, ctx?: { database?: string; schema?: string }) => void
}) {
  const { t } = useTranslation()
  const { data } = useQuery({
    queryKey: ['tables', connectionId, database, schema],
    queryFn: () => schemaApi.listTables(connectionId, { database, schema }),
  })
  const tables = data?.items ?? []
  if (tables.length === 0)
    return <p className="ml-4 text-xs text-muted">{t('connection.emptyTables')}</p>
  return (
    <ul className="ml-4">
      {tables.map((tbl) => (
        <li key={tbl.name}>
          <button
            type="button"
            className="w-full truncate text-left text-xs hover:text-accent"
            onClick={() => onSelectTable(tbl.name, { database, schema })}
          >
            {tbl.type === 'view' ? '👁' : '📋'} {tbl.name}
          </button>
        </li>
      ))}
    </ul>
  )
}
