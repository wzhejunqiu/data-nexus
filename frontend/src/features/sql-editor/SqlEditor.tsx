import Editor from '@monaco-editor/react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { connectionApi } from '@/lib/api/connection'
import { queryApi } from '@/lib/api/query'
import type { QueryResponse } from '@/lib/types'
import { formatCell, isWriteSQL, rowsToCSV } from '@/lib/utils'
import { useQueryHistoryStore } from '@/stores/queryHistoryStore'

const PRAGMAS = [
  'PRAGMA table_info(sqlite_master);',
  'PRAGMA foreign_keys;',
  'PRAGMA journal_mode;',
  'PRAGMA user_version;',
]

export function SqlEditor({ connectionId }: { connectionId: string | null }) {
  const { t } = useTranslation()
  const [sql, setSql] = useState('SELECT 1;')
  const [result, setResult] = useState<QueryResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const addHistory = useQueryHistoryStore((s) => s.add)
  const history = useQueryHistoryStore((s) => (connectionId ? (s.items[connectionId] ?? []) : []))

  const { data: connections } = useQuery({
    queryKey: ['connections'],
    queryFn: () => connectionApi.list(),
  })

  const openConnections = connections?.items.filter((c) => c.status === 'open') ?? []
  const [activeConn, setActiveConn] = useState(connectionId ?? '')

  useEffect(() => {
    if (connectionId) setActiveConn(connectionId)
  }, [connectionId])

  const activeConnItem = openConnections.find((c) => c.id === activeConn)
  const connReadOnly = activeConnItem?.config.sqlite?.readOnly ?? false

  const execute = useMutation({
    mutationFn: async () => {
      if (!activeConn) throw new Error('no connection')
      if (connReadOnly && isWriteSQL(sql)) throw new Error(t('sql.readOnlyBlocked'))
      if (isWriteSQL(sql) && !window.confirm(t('sql.confirmWrite'))) {
        throw new Error('cancelled')
      }
      return queryApi.execute({ connectionId: activeConn, sql, maxRows: 1000 })
    },
    onSuccess: (res) => {
      setError(null)
      setResult(res)
      addHistory(activeConn, sql)
    },
    onError: (err) => {
      if ((err as Error).message === 'cancelled') return
      setError((err as Error).message)
      setResult(null)
    },
  })

  const copyCsv = () => {
    if (!result?.columns || !result.rows) return
    const csv = rowsToCSV(
      result.columns.map((c) => c.name),
      result.rows,
    )
    navigator.clipboard.writeText(csv)
  }

  return (
    <div className="flex h-full flex-col gap-3 p-4">
      <div className="flex flex-wrap items-center gap-2">
        <label className="text-sm">
          {t('sql.connection')}
          <select
            className="ml-2 rounded border border-border bg-transparent px-2 py-1 text-sm"
            value={activeConn}
            onChange={(e) => setActiveConn(e.target.value)}
          >
            <option value="">—</option>
            {openConnections.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </label>
        <Button onClick={() => execute.mutate()} disabled={!activeConn || execute.isPending}>
          {t('sql.run')} (⌘↵)
        </Button>
        {history.length > 0 && (
          <select
            className="rounded border border-border bg-transparent px-2 py-1 text-sm"
            onChange={(e) => e.target.value && setSql(e.target.value)}
            defaultValue=""
          >
            <option value="">{t('sql.history')}</option>
            {history.map((h) => (
              <option key={h} value={h}>
                {h.slice(0, 60)}
              </option>
            ))}
          </select>
        )}
        <select
          className="rounded border border-border bg-transparent px-2 py-1 text-sm"
          onChange={(e) => e.target.value && setSql(e.target.value)}
          defaultValue=""
        >
          <option value="">{t('sql.pragma')}</option>
          {PRAGMAS.map((p) => (
            <option key={p} value={p}>
              {p}
            </option>
          ))}
        </select>
      </div>
      <div
        className="overflow-hidden rounded border border-border"
        onKeyDown={(e) => {
          if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
            e.preventDefault()
            execute.mutate()
          }
        }}
      >
        <Editor
          height="180px"
          defaultLanguage="sql"
          theme="vs-dark"
          value={sql}
          onChange={(v) => setSql(v ?? '')}
          options={{ minimap: { enabled: false }, fontSize: 13, lineNumbers: 'on' }}
        />
      </div>
      {error && <p className="rounded bg-red-500/10 p-2 text-sm text-red-500">{error}</p>}
      {result && (
        <div className="min-h-0 flex-1 space-y-2 overflow-auto">
          <div className="flex items-center gap-3 text-sm text-muted">
            <span>{t('sql.duration', { ms: result.durationMs })}</span>
            {result.kind === 'exec' && (
              <span>{t('sql.rowsAffected', { count: result.rowsAffected ?? 0 })}</span>
            )}
            {result.kind === 'result' && result.columns && (
              <Button size="sm" variant="outline" onClick={copyCsv}>
                {t('sql.copyCsv')}
              </Button>
            )}
          </div>
          {result.kind === 'result' && result.columns && (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  {result.columns.map((c) => (
                    <th key={c.name} className="px-2 py-1 text-left">
                      {c.name}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {(result.rows ?? []).map((row, i) => (
                  <tr key={i} className="border-b border-border/40">
                    {result.columns!.map((c) => (
                      <td key={c.name} className="px-2 py-1 font-mono text-xs">
                        {row[c.name] === null || row[c.name] === undefined
                          ? t('data.null')
                          : formatCell(row[c.name])}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  )
}
