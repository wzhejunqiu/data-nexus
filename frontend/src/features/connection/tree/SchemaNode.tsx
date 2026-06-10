import { useQuery } from '@tanstack/react-query'
import { schemaApi } from '@/lib/api/schema'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import { TreeChevron } from './TreeNodeRow'
import { NamespaceTables } from './namespaceTables'

export function SchemaList({
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
        return (
          <SchemaNode
            key={key}
            connectionId={connectionId}
            database={database}
            schemaName={sch.name}
            open={open}
            searchQuery={q}
            onToggleExpand={() => setExpandedSchema((s) => ({ ...s, [key]: !open }))}
            onExpand={() => setExpandedSchema((s) => ({ ...s, [key]: true }))}
            onSelectTable={onSelectTable}
          />
        )
      })}
    </div>
  )
}

export function SchemaNode({
  connectionId,
  database,
  schemaName,
  open,
  searchQuery,
  onToggleExpand,
  onExpand,
  onSelectTable,
}: {
  connectionId: string
  database: string
  schemaName: string
  open: boolean
  searchQuery: string
  onToggleExpand: () => void
  onExpand: () => void
  onSelectTable: (table: string, ctx?: { database?: string; schema?: string }) => void
}) {
  const setBrowseContext = useWorkspaceStore((s) => s.setBrowseContext)
  const browseDatabase = useWorkspaceStore((s) => s.browseContext[connectionId]?.database)
  const browseSchema = useWorkspaceStore((s) => s.browseContext[connectionId]?.schema)
  const isActiveBrowse = browseDatabase === database && browseSchema === schemaName

  const setBrowse = () => setBrowseContext(connectionId, { database, schema: schemaName })

  return (
    <div>
      <div
        className={`flex w-full items-center gap-1 text-left text-xs text-muted hover:text-foreground ${
          isActiveBrowse ? 'rounded ring-1 ring-accent/50' : ''
        }`}
        onClick={() => setBrowse()}
        onDoubleClick={() => {
          setBrowse()
          onExpand()
        }}
      >
        <TreeChevron
          open={open}
          onToggle={(e) => {
            e.stopPropagation()
            onToggleExpand()
          }}
        />
        <span>{schemaName}</span>
      </div>
      {open && (
        <NamespaceTables
          connectionId={connectionId}
          database={database}
          schema={schemaName}
          onSelectTable={onSelectTable}
          searchQuery={searchQuery}
        />
      )}
    </div>
  )
}
