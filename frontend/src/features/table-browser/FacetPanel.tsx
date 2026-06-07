import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { schemaApi } from '@/lib/api/schema'
import type { RowFilter } from '@/lib/types'

export function FacetPanel({
  connectionId,
  tableName,
  filters,
  onAddFilter,
}: {
  connectionId: string
  tableName: string
  filters: RowFilter[]
  onAddFilter: (filter: RowFilter) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const { data } = useQuery({
    queryKey: ['profile', connectionId, tableName],
    queryFn: () => schemaApi.getTableProfile(connectionId, tableName),
  })

  const facetColumns = data?.columns.filter((c) => c.isLowCardinality && c.topValues?.length) ?? []
  if (facetColumns.length === 0) return null

  const activeValues = new Set(
    filters.filter((f) => f.operator === 'eq').map((f) => `${f.column}=${f.value}`),
  )

  return (
    <div className="mb-3 rounded border border-border p-3">
      <button
        type="button"
        className="text-xs font-semibold text-muted hover:text-foreground"
        onClick={() => setOpen((v) => !v)}
      >
        {open ? '▼' : '▶'} {t('facet.title')} ({facetColumns.length})
      </button>
      {open && (
        <div className="mt-2 space-y-3">
          {facetColumns.map((col) => (
            <div key={col.name}>
              <p className="text-xs font-medium">{col.name}</p>
              <div className="mt-1 flex flex-wrap gap-1">
                {col.topValues?.map((tv) => {
                  const key = `${col.name}=${tv.value}`
                  const active = activeValues.has(key)
                  return (
                    <button
                      key={tv.value}
                      type="button"
                      className={`rounded px-2 py-0.5 text-xs ${
                        active ? 'bg-accent/30 text-accent' : 'bg-muted/30 hover:bg-muted/50'
                      }`}
                      onClick={() =>
                        onAddFilter({ column: col.name, operator: 'eq', value: tv.value })
                      }
                    >
                      {tv.value} ({tv.count})
                    </button>
                  )
                })}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
