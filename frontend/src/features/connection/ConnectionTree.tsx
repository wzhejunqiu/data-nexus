import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { useToastStore } from '@/components/ui/Toast'
import { connectionApi } from '@/lib/api/connection'
import { formatError } from '@/lib/api/errors'
import type { ConnectionListItem } from '@/lib/types'
import { SchemaSubtree } from '@/features/schema/SchemaSubtree'
import { NewConnectionDialog } from './NewConnectionDialog'
import { EditConnectionDialog } from './EditConnectionDialog'
import { RenameDialog } from '@/features/schema/IndexList'
import { useWorkspaceStore } from '@/stores/workspaceStore'

export function ConnectionTree() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [error, setError] = useState<string | null>(null)
  const [readOnly, setReadOnly] = useState(false)
  const [wal, setWal] = useState(false)
  const [newOpen, setNewOpen] = useState(false)
  const [sidebarWidth, setSidebarWidth] = useState(260)
  const selectTable = useWorkspaceStore((s) => s.selectTable)
  const setActiveConnectionId = useWorkspaceStore((s) => s.setActiveConnectionId)

  const { data, isLoading } = useQuery({
    queryKey: ['connections'],
    queryFn: () => connectionApi.list(),
  })

  const refresh = () => qc.invalidateQueries({ queryKey: ['connections'] })

  const openConn = useMutation({
    mutationFn: (id: string) => connectionApi.open(id),
    onSuccess: (conn) => {
      setError(null)
      setActiveConnectionId(conn.id)
      refresh()
      qc.invalidateQueries({ queryKey: ['tables', conn.id] })
    },
    onError: (err) => setError(formatError(t, err)),
  })

  const closeConn = useMutation({
    mutationFn: (id: string) => connectionApi.close(id),
    onSuccess: (_, id) => {
      const { activeConnectionId, setSelectedTable } = useWorkspaceStore.getState()
      if (activeConnectionId === id) {
        setActiveConnectionId(null)
        setSelectedTable(null)
      }
      refresh()
    },
    onError: (err) => setError(formatError(t, err)),
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
      setError(formatError(t, err))
    },
  })

  const items = data?.items ?? []
  const openCount = items.filter((i) => i.status === 'open').length

  return (
    <>
      <aside
        className="relative flex flex-col border-r border-border"
        style={{ width: sidebarWidth, minWidth: 200, maxWidth: 480 }}
      >
        <div className="flex flex-col gap-2 border-b border-border p-3">
          <Button onClick={() => setNewOpen(true)}>{t('connection.new')}</Button>
          <label className="flex items-center gap-2 text-xs text-muted">
            <input
              type="checkbox"
              checked={readOnly}
              onChange={(e) => {
                const next = e.target.checked
                setReadOnly(next)
                if (next) setWal(false)
              }}
            />
            {t('connection.readOnly')}
          </label>
          <label className="flex items-center gap-2 text-xs text-muted">
            <input
              type="checkbox"
              checked={wal}
              disabled={readOnly}
              onChange={(e) => setWal(e.target.checked)}
            />
            {t('connection.wal')}
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
              onRenamed={refresh}
            />
          ))}
        </div>
        <div className="border-t border-border p-2 text-xs text-muted">
          {t('app.openCount', { count: openCount })}
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
      <NewConnectionDialog
        open={newOpen}
        onOpenChange={setNewOpen}
        readOnly={readOnly}
        wal={wal}
      />
    </>
  )
}

function RestoreOnStartupToggle() {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const [enabled, setEnabled] = useState(false)
  const [loaded, setLoaded] = useState(false)

  useEffect(() => {
    connectionApi
      .getRestoreOpenOnStartup()
      .then(setEnabled)
      .catch((err) => pushToast(formatError(t, err), 'error'))
      .finally(() => setLoaded(true))
  }, [pushToast, t])

  if (!loaded) return null

  return (
    <label className="flex items-center gap-2 text-xs text-muted">
      <input
        type="checkbox"
        checked={enabled}
        onChange={async (e) => {
          const next = e.target.checked
          setEnabled(next)
          try {
            await connectionApi.setRestoreOpenOnStartup(next)
          } catch (err) {
            setEnabled(!next)
            pushToast(formatError(t, err), 'error')
          }
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
  onRenamed,
}: {
  item: ConnectionListItem
  onOpen: () => void
  onClose: () => void
  onRemove: () => void
  onSelectTable: (table: string) => void
  onRenamed: () => void
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const isOpen = item.status === 'open'
  const path = item.config.sqlite?.filePath ?? ''
  const [renameOpen, setRenameOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [name, setName] = useState(item.name)

  const rename = async () => {
    try {
      await connectionApi.rename(item.id, name.trim() || item.name)
      setRenameOpen(false)
      onRenamed()
    } catch (err) {
      pushToast(formatError(t, err), 'error')
    }
  }

  return (
    <div
      className="mb-2 rounded-md border border-border/60 p-2"
      onDoubleClick={() => {
        if (!isOpen) onOpen()
      }}
    >
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
          <Button
            size="sm"
            variant="outline"
            onClick={() => {
              setRenameOpen(true)
            }}
          >
            {t('connection.rename')}
          </Button>
          <Button size="sm" variant="outline" onClick={() => setEditOpen(true)}>
            {t('connection.edit')}
          </Button>
          <Button size="sm" variant="danger" onClick={onRemove}>
            {t('connection.remove')}
          </Button>
        </div>
      </div>
      {isOpen && <SchemaSubtree connectionId={item.id} onSelectTable={onSelectTable} />}
      <RenameDialog
        open={renameOpen}
        name={name}
        onNameChange={setName}
        onOpenChange={(open) => {
          setRenameOpen(open)
          if (open) setName(item.name)
        }}
        onConfirm={rename}
      />
      <EditConnectionDialog
        item={item}
        open={editOpen}
        onOpenChange={setEditOpen}
        onSaved={onRenamed}
      />
    </div>
  )
}
