import Editor, { type OnMount } from '@monaco-editor/react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { connectionApi } from '@/lib/api/connection'
import { formatError } from '@/lib/api/errors'
import { queryApi } from '@/lib/api/query'
import type { QueryResponse } from '@/lib/types'
import { formatCell, isWriteSQL, rowsToCSV } from '@/lib/utils'
import { useQueryHistoryStore } from '@/stores/queryHistoryStore'
import { resolveTheme, useThemeStore } from '@/stores/themeStore'
import { useStatusStore } from '@/stores/statusStore'
import { ConfirmDialog } from './ConfirmDialog'
import { QueryHistory } from './QueryHistory'

const PRAGMAS = [
  'PRAGMA table_info(sqlite_master);',
  'PRAGMA foreign_keys;',
  'PRAGMA journal_mode;',
  'PRAGMA user_version;',
]

const EMPTY_HISTORY: string[] = []

export function SqlEditor({ connectionId }: { connectionId: string | null }) {
  const { t } = useTranslation()
  const themeMode = useThemeStore((s) => s.mode)
  const monacoTheme = resolveTheme(themeMode) === 'dark' ? 'vs-dark' : 'vs'
  const setStatus = useStatusStore((s) => s.setStatus)
  const [sql, setSql] = useState('SELECT 1;')
  const [result, setResult] = useState<QueryResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingRun, setPendingRun] = useState(false)
  const addHistory = useQueryHistoryStore((s) => s.add)
  const history = useQueryHistoryStore((s) =>
    connectionId ? (s.items[connectionId] ?? EMPTY_HISTORY) : EMPTY_HISTORY,
  )

  const { data: connections } = useQuery({
    queryKey: ['connections'],
    queryFn: () => connectionApi.list(),
  })

  const openConnections = connections?.items.filter((c) => c.status === 'open') ?? []
  const [selectedConn, setSelectedConn] = useState('')
  const activeConn = connectionId || selectedConn

  const activeConnItem = openConnections.find((c) => c.id === activeConn)
  const connReadOnly = activeConnItem?.config.sqlite?.readOnly ?? false

  const runQuery = useCallback(async () => {
    if (!activeConn) throw new Error(t('sql.noConnection'))
    const res = await queryApi.execute({ connectionId: activeConn, sql, maxRows: 1000 })
    setError(null)
    setResult(res)
    addHistory(activeConn, sql)
    if (res.kind === 'result') {
      setStatus('query', res.rowCount ?? res.rows?.length ?? 0, res.durationMs)
    } else {
      setStatus('exec', res.rowsAffected ?? 0, res.durationMs)
    }
    return res
  }, [activeConn, sql, addHistory, setStatus, t])

  const execute = useMutation({
    mutationFn: runQuery,
    onError: (err) => {
      setError(formatError(t, err))
      setResult(null)
    },
  })

  const handleRunRef = useRef<() => void>(() => {})

  const handleRun = () => {
    if (!activeConn) return
    if (connReadOnly && isWriteSQL(sql)) {
      setError(t('sql.readOnlyBlocked'))
      return
    }
    if (isWriteSQL(sql)) {
      setConfirmOpen(true)
      setPendingRun(true)
      return
    }
    execute.mutate()
  }

  useEffect(() => {
    handleRunRef.current = handleRun
  })

  const handleEditorMount: OnMount = (editor, monaco) => {
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
      handleRunRef.current()
    })
  }

  const resultColumns = useMemo(() => result?.columns?.map((c) => c.name) ?? [], [result])

  const copyCsv = () => {
    if (!result?.columns || !result.rows) return
    navigator.clipboard.writeText(rowsToCSV(resultColumns, result.rows))
  }

  return (
    <div className="flex h-full flex-col gap-3 p-4">
      <div className="flex flex-wrap items-center gap-2">
        <label className="text-sm">
          {t('sql.connection')}
          <select
            className="ml-2 rounded border border-border bg-transparent px-2 py-1 text-sm"
            value={activeConn}
            onChange={(e) => setSelectedConn(e.target.value)}
          >
            <option value="">—</option>
            {openConnections.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </label>
        <Button onClick={handleRun} disabled={!activeConn || execute.isPending}>
          {t('sql.run')} (⌘↵)
        </Button>
        <QueryHistory history={history} onSelect={setSql} />
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
      <div className="overflow-hidden rounded border border-border">
        <Editor
          height="180px"
          defaultLanguage="sql"
          theme={monacoTheme}
          value={sql}
          onChange={(v) => setSql(v ?? '')}
          onMount={handleEditorMount}
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
            {result.kind === 'result' && (
              <span>
                {t('sql.rowCount', { count: result.rowCount ?? result.rows?.length ?? 0 })}
              </span>
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
      <ConfirmDialog
        open={confirmOpen}
        message={t('sql.confirmWrite')}
        onOpenChange={(open) => {
          setConfirmOpen(open)
          if (!open) setPendingRun(false)
        }}
        onConfirm={() => {
          setConfirmOpen(false)
          if (pendingRun) execute.mutate()
          setPendingRun(false)
        }}
      />
    </div>
  )
}
