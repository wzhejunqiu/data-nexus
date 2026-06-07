import { useQuery } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { schemaApi } from '@/lib/api/schema'
import { useWorkspaceStore } from '@/stores/workspaceStore'

export function SchemaSubtree({
  connectionId,
  onSelectTable,
}: {
  connectionId: string
  onSelectTable: (table: string) => void
}) {
  const { t } = useTranslation()
  const [filter, setFilter] = useState('')
  const selectedTable = useWorkspaceStore((s) => s.selectedTable)
  const activeConnectionId = useWorkspaceStore((s) => s.activeConnectionId)

  const { data, isLoading } = useQuery({
    queryKey: ['tables', connectionId],
    queryFn: () => schemaApi.listTables(connectionId),
    enabled: !!connectionId,
  })

  const items = useMemo(() => {
    const list = data?.items ?? []
    if (!filter) return list
    const q = filter.toLowerCase()
    return list.filter((i) => i.name.toLowerCase().includes(q))
  }, [data, filter])

  const tables = items.filter((i) => i.type === 'table')
  const views = items.filter((i) => i.type === 'view')

  if (isLoading) return <p className="mt-2 text-xs text-muted">{t('common.loading')}</p>

  return (
    <div className="mt-2 space-y-2">
      <input
        className="w-full rounded border border-border bg-transparent px-2 py-1 text-xs"
        placeholder={t('connection.search')}
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
      />
      {items.length === 0 && <p className="text-xs text-muted">{t('schema.empty')}</p>}
      {tables.length > 0 && (
        <SchemaGroup
          title={t('connection.tables')}
          names={tables.map((t) => t.name)}
          selected={activeConnectionId === connectionId ? selectedTable : null}
          onSelect={onSelectTable}
        />
      )}
      {views.length > 0 && (
        <SchemaGroup
          title={t('connection.views')}
          names={views.map((v) => v.name)}
          selected={activeConnectionId === connectionId ? selectedTable : null}
          onSelect={onSelectTable}
        />
      )}
    </div>
  )
}

function SchemaGroup({
  title,
  names,
  selected,
  onSelect,
}: {
  title: string
  names: string[]
  selected: string | null
  onSelect: (name: string) => void
}) {
  return (
    <div>
      <p className="text-xs font-semibold text-muted">{title}</p>
      <ul className="mt-1 space-y-0.5">
        {names.map((name) => (
          <li key={name}>
            <button
              className={`w-full truncate rounded px-2 py-1 text-left text-xs hover:bg-muted/40 ${
                selected === name ? 'bg-accent/20 text-accent' : ''
              }`}
              onClick={() => onSelect(name)}
            >
              {name}
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
