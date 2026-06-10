import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { schemaApi } from '@/lib/api/schema'
import { tableKey } from '@/lib/tableKey'
import type { TableInfo } from '@/lib/types'

export function NamespaceTables({
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
