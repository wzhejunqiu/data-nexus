import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { connectionApi } from '@/lib/api/connection'
import { dialogApi } from '@/lib/api/dialog'
import { schemaApi } from '@/lib/api/schema'
import { formatError } from '@/lib/api/errors'
import { tableKey } from '@/lib/tableKey'
import type { TableInfo } from '@/lib/types'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import { useToastStore } from '@/components/ui/Toast'

export function SchemaSubtree({
  connectionId,
  onSelectTable,
}: {
  connectionId: string
  onSelectTable: (table: string) => void
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const pushToast = useToastStore((s) => s.push)
  const [filter, setFilter] = useState('')
  const selectedTable = useWorkspaceStore((s) => s.selectedTable)
  const activeConnectionId = useWorkspaceStore((s) => s.activeConnectionId)

  const { data, isLoading } = useQuery({
    queryKey: ['tables', connectionId],
    queryFn: () => schemaApi.listTables(connectionId),
    enabled: !!connectionId,
  })

  const { data: attached } = useQuery({
    queryKey: ['attached', connectionId],
    queryFn: () => connectionApi.listAttached(connectionId),
    enabled: !!connectionId,
  })

  const attachDb = useMutation({
    mutationFn: async () => {
      const path = await dialogApi.openDatabaseFile()
      const base =
        path
          .split(/[/\\]/)
          .pop()
          ?.replace(/\.[^.]+$/, '') ?? 'db'
      const alias = window.prompt(t('attach.aliasPrompt'), base)
      if (!alias?.trim()) throw new Error('cancelled')
      await connectionApi.attach(connectionId, path, alias.trim())
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tables', connectionId] })
      qc.invalidateQueries({ queryKey: ['attached', connectionId] })
    },
    onError: (err) => {
      if ((err as Error).message === 'cancelled') return
      pushToast(formatError(t, err), 'error')
    },
  })

  const detachDb = useMutation({
    mutationFn: (alias: string) => connectionApi.detach(connectionId, alias),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tables', connectionId] })
      qc.invalidateQueries({ queryKey: ['attached', connectionId] })
    },
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

  const items = useMemo(() => {
    const list = data?.items ?? []
    if (!filter) return list
    const q = filter.toLowerCase()
    return list.filter((i) => i.name.toLowerCase().includes(q))
  }, [data, filter])

  const groups = useMemo(() => {
    const map = new Map<string, TableInfo[]>()
    for (const item of items) {
      const schema = item.schema ?? 'main'
      const list = map.get(schema) ?? []
      list.push(item)
      map.set(schema, list)
    }
    return [...map.entries()].sort(([a], [b]) => a.localeCompare(b))
  }, [items])

  if (isLoading) return <p className="mt-2 text-xs text-muted">{t('common.loading')}</p>

  return (
    <div className="mt-2 space-y-2">
      <div className="flex gap-1">
        <Button
          size="sm"
          variant="outline"
          className="flex-1 text-xs"
          onClick={() => attachDb.mutate()}
        >
          {t('attach.attach')}
        </Button>
      </div>
      {(attached ?? []).length > 0 && (
        <ul className="space-y-0.5 text-xs">
          {(attached ?? []).map((a) => (
            <li
              key={a.alias}
              className="flex items-center justify-between rounded bg-muted/20 px-2 py-1"
            >
              <span className="truncate font-mono">{a.alias}</span>
              <button
                type="button"
                className="text-muted hover:text-red-500"
                onClick={() => detachDb.mutate(a.alias)}
              >
                {t('attach.detach')}
              </button>
            </li>
          ))}
        </ul>
      )}
      <input
        className="w-full rounded border border-border bg-transparent px-2 py-1 text-xs"
        placeholder={t('connection.search')}
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
      />
      {items.length === 0 && <p className="text-xs text-muted">{t('schema.empty')}</p>}
      {groups.map(([schema, schemaItems]) => {
        const tables = schemaItems.filter((i) => i.type === 'table')
        const views = schemaItems.filter((i) => i.type === 'view')
        const label = schema === 'main' ? t('attach.mainDb') : schema
        return (
          <div key={schema}>
            <p className="text-xs font-bold text-muted">{label}</p>
            {tables.length > 0 && (
              <SchemaGroup
                title={t('connection.tables')}
                items={tables}
                selected={activeConnectionId === connectionId ? selectedTable : null}
                onSelect={(item) => onSelectTable(tableKey(item))}
              />
            )}
            {views.length > 0 && (
              <SchemaGroup
                title={t('connection.views')}
                items={views}
                selected={activeConnectionId === connectionId ? selectedTable : null}
                onSelect={(item) => onSelectTable(tableKey(item))}
              />
            )}
          </div>
        )
      })}
    </div>
  )
}

function SchemaGroup({
  title,
  items,
  selected,
  onSelect,
}: {
  title: string
  items: TableInfo[]
  selected: string | null
  onSelect: (item: TableInfo) => void
}) {
  const { t } = useTranslation()

  return (
    <div>
      <p className="text-xs font-semibold text-muted">{title}</p>
      <ul className="mt-1 space-y-0.5">
        {items.map((item) => {
          const key = tableKey(item)
          const isSelected = selected === key
          return (
            <li key={key}>
              <button
                type="button"
                aria-current={isSelected ? 'true' : undefined}
                className={`flex w-full items-center justify-between gap-2 truncate rounded px-2 py-1 text-left text-xs transition-colors ${
                  isSelected ? 'bg-accent/20 text-accent hover:bg-accent/30' : 'hover:bg-muted/40'
                }`}
                onClick={() => onSelect(item)}
              >
                <span className="truncate">{item.name}</span>
                {item.type === 'table' && item.rowCount != null && (
                  <span className="shrink-0 text-muted">
                    {t('data.rowCountShort', { count: item.rowCount })}
                  </span>
                )}
              </button>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
