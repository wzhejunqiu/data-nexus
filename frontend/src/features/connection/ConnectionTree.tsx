import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { connectionApi } from '@/lib/api/connection'
import { dialogApi } from '@/lib/api/dialog'
import { isDialogCancelled } from '@/lib/api/errors'
import type { AppError, ConnectionListItem } from '@/lib/types'
import { SchemaSubtree } from '@/features/schema/SchemaSubtree'
import { useWorkspaceStore } from '@/stores/workspaceStore'

export function ConnectionTree() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [error, setError] = useState<string | null>(null)
  const [readOnly, setReadOnly] = useState(false)
  const selectTable = useWorkspaceStore((s) => s.selectTable)

  const { data, isLoading } = useQuery({
    queryKey: ['connections'],
    queryFn: () => connectionApi.list(),
  })

  const refresh = () => qc.invalidateQueries({ queryKey: ['connections'] })

  const openFromFile = useMutation({
    mutationFn: async () => {
      const path = await dialogApi.openDatabaseFile()
      return connectionApi.openFromFile({ filePath: path, readOnly })
    },
    onSuccess: (conn) => {
      setError(null)
      refresh()
      qc.invalidateQueries({ queryKey: ['tables', conn.id] })
    },
    onError: (err) => {
      const appErr = err as unknown as AppError
      if (isDialogCancelled(appErr)) return
      setError(appErr.message ?? String(err))
    },
  })

  const openConn = useMutation({
    mutationFn: (id: string) => connectionApi.open(id),
    onSuccess: (conn) => {
      setError(null)
      refresh()
      qc.invalidateQueries({ queryKey: ['tables', conn.id] })
    },
    onError: (err) => setError((err as { message: string }).message),
  })

  const closeConn = useMutation({
    mutationFn: (id: string) => connectionApi.close(id),
    onSuccess: refresh,
    onError: (err) => setError((err as { message: string }).message),
  })

  const removeConn = useMutation({
    mutationFn: (item: ConnectionListItem) => {
      const msg =
        item.status === 'open'
          ? t('connection.confirmRemoveOpen', { name: item.name })
          : t('connection.confirmRemove', { name: item.name })
      if (!window.confirm(msg)) return Promise.reject(new Error('cancelled'))
      return connectionApi.remove(item.id)
    },
    onSuccess: refresh,
    onError: (err) => {
      if ((err as Error).message === 'cancelled') return
      setError((err as { message: string }).message)
    },
  })

  const items = data?.items ?? []
  const openCount = items.filter((i) => i.status === 'open').length

  return (
    <aside className="flex w-72 flex-col border-r border-border">
      <div className="flex flex-col gap-2 border-b border-border p-3">
        <Button onClick={() => openFromFile.mutate()} disabled={openFromFile.isPending}>
          {t('connection.new')}
        </Button>
        <label className="flex items-center gap-2 text-xs text-muted">
          <input
            type="checkbox"
            checked={readOnly}
            onChange={(e) => setReadOnly(e.target.checked)}
          />
          {t('connection.readOnly')}
        </label>
        <RestoreOnStartupToggle />
        {error && <p className="text-xs text-red-500">{error}</p>}
      </div>
      <div className="flex-1 overflow-auto p-2">
        {isLoading && <p className="text-sm text-muted">{t('common.loading')}</p>}
        {!isLoading && items.length === 0 && (
          <p className="text-sm text-muted">{t('app.noConnection')}</p>
        )}
        {items.map((item) => (
          <ConnectionTreeItem
            key={item.id}
            item={item}
            onOpen={() => openConn.mutate(item.id)}
            onClose={() => closeConn.mutate(item.id)}
            onRemove={() => removeConn.mutate(item)}
            onSelectTable={(table) => selectTable(item.id, table)}
          />
        ))}
      </div>
      <div className="border-t border-border p-2 text-xs text-muted">
        {t('app.openCount', { count: openCount })}
      </div>
    </aside>
  )
}

function RestoreOnStartupToggle() {
  const { t } = useTranslation()
  const [enabled, setEnabled] = useState(false)

  return (
    <label className="flex items-center gap-2 text-xs text-muted">
      <input
        type="checkbox"
        checked={enabled}
        onChange={async (e) => {
          const next = e.target.checked
          setEnabled(next)
          await connectionApi.setRestoreOpenOnStartup(next)
        }}
      />
      {t('connection.restoreOnStartup')}
    </label>
  )
}

function ConnectionTreeItem({
  item,
  onOpen,
  onClose,
  onRemove,
  onSelectTable,
}: {
  item: ConnectionListItem
  onOpen: () => void
  onClose: () => void
  onRemove: () => void
  onSelectTable: (table: string) => void
}) {
  const { t } = useTranslation()
  const isOpen = item.status === 'open'
  const path = item.config.sqlite?.filePath ?? ''

  return (
    <div className="mb-2 rounded-md border border-border/60 p-2">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1 text-sm font-medium">
            <span className={isOpen ? 'text-green-500' : 'text-muted'}>{isOpen ? '●' : '○'}</span>
            <span className="truncate">{item.name}</span>
          </div>
          <p className="truncate text-xs text-muted" title={path}>
            {path}
          </p>
        </div>
        <div className="flex flex-col gap-1">
          {isOpen ? (
            <Button size="sm" variant="outline" onClick={onClose}>
              {t('connection.close')}
            </Button>
          ) : (
            <Button size="sm" onClick={onOpen}>
              {t('connection.open')}
            </Button>
          )}
          <Button size="sm" variant="danger" onClick={onRemove}>
            {t('connection.remove')}
          </Button>
        </div>
      </div>
      {isOpen && <SchemaSubtree connectionId={item.id} onSelectTable={onSelectTable} />}
    </div>
  )
}
