import Editor, { type OnMount } from '@monaco-editor/react'
import type { editor, languages, Position } from 'monaco-editor'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { NullCell } from '@/components/ui/NullCell'
import { useToastStore } from '@/components/ui/Toast'
import { useExportStore } from '@/stores/exportStore'
import { connectionApi } from '@/lib/api/connection'
import { formatError } from '@/lib/api/errors'
import { connectionTypeToSqlDialect } from '@/lib/sql/dialect'
import { formatSql } from '@/lib/sql/format'
import { queryApi } from '@/lib/api/query'
import { schemaApi } from '@/lib/api/schema'
import type { QueryResponse } from '@/lib/types'
import { loadCSVFormatPreference } from '@/lib/types/csv'
import { formatCell, rowsToCSV } from '@/lib/utils'
import { useQueryHistoryStore } from '@/stores/queryHistoryStore'
import { resolveTheme, useThemeStore } from '@/stores/themeStore'
import { useStatusStore } from '@/stores/statusStore'
import { ConfirmDialog } from './ConfirmDialog'
import { ExplainTree, isExplainResult } from './ExplainTree'
import { QueryHistory } from './QueryHistory'

const PRAGMAS = [
  'PRAGMA table_info(sqlite_master);',
  'PRAGMA foreign_keys;',
  'PRAGMA journal_mode;',
  'PRAGMA user_version;',
]

const SQL_KEYWORDS = [
  'SELECT', 'FROM', 'WHERE', 'JOIN', 'LEFT', 'RIGHT', 'INNER', 'OUTER', 'ON', 'GROUP', 'BY',
  'ORDER', 'HAVING', 'LIMIT', 'INSERT', 'INTO', 'VALUES', 'UPDATE', 'SET', 'DELETE', 'CREATE',
  'TABLE', 'INDEX', 'AND', 'OR', 'NOT', 'NULL', 'AS', 'DISTINCT', 'EXPLAIN', 'WITH', 'PRAGMA',
]

const EMPTY_HISTORY: string[] = []
const MIN_EDITOR_LINES = 8
const MAX_EDITOR_LINES = 20
const EDITOR_LINE_HEIGHT = 19

export function SqlEditor({ connectionId }: { connectionId: string | null }) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const themeMode = useThemeStore((s) => s.mode)
  const monacoTheme = resolveTheme(themeMode) === 'dark' ? 'vs-dark' : 'vs'
  const setStatus = useStatusStore((s) => s.setStatus)
  const [sql, setSql] = useState('SELECT 1;')
  const [result, setResult] = useState<QueryResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingRun, setPendingRun] = useState(false)
  const openExport = useExportStore((s) => s.openExport)
  const [editorHeight, setEditorHeight] = useState(MIN_EDITOR_LINES * EDITOR_LINE_HEIGHT)
  const editorRef = useRef<Parameters<OnMount>[0] | null>(null)
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

  const { data: tables } = useQuery({
    queryKey: ['schema', 'tables', activeConn],
    queryFn: () => schemaApi.listTables(activeConn),
    enabled: !!activeConn,
  })

  const schemaCache = useRef<Map<string, string[]>>(new Map())

  const activeConnItem = openConnections.find((c) => c.id === activeConn)
  const connReadOnly = activeConnItem?.config.sqlite?.readOnly ?? false

  const runQuery = useCallback(async () => {
    if (!activeConn) throw new Error(t('sql.noConnection'))
    const res = await queryApi.execute({ connectionId: activeConn, sql, maxRows: 10000 })
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

  const handleRun = async () => {
    if (!activeConn) return
    try {
      const kind = await queryApi.classifySQL(activeConn, sql)
      if (connReadOnly && kind === 'write') {
        setError(t('sql.readOnlyBlocked'))
        return
      }
      if (kind === 'write') {
        setConfirmOpen(true)
        setPendingRun(true)
        return
      }
      setError(null)
      execute.mutate()
    } catch (err) {
      setError(formatError(t, err))
    }
  }

  useEffect(() => {
    handleRunRef.current = () => {
      void handleRun()
    }
  })

  const tableNames = useMemo(() => tables?.items.map((t) => t.name) ?? [], [tables])

  const handleEditorMount: OnMount = (editor, monaco) => {
    editorRef.current = editor
    const updateHeight = () => {
      const lines = Math.max(MIN_EDITOR_LINES, editor.getContentHeight() / EDITOR_LINE_HEIGHT)
      setEditorHeight(Math.min(MAX_EDITOR_LINES, lines) * EDITOR_LINE_HEIGHT)
    }
    updateHeight()
    editor.onDidContentSizeChange(updateHeight)
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
      handleRunRef.current()
    })
    editor.addCommand(monaco.KeyMod.Shift | monaco.KeyMod.Alt | monaco.KeyCode.KeyF, () => {
      formatEditorSQL()
    })

    monaco.languages.registerCompletionItemProvider('sql', {
      triggerCharacters: ['.', ' '],
      provideCompletionItems: async (model: editor.ITextModel, position: Position) => {
        const word = model.getWordUntilPosition(position)
        const range = {
          startLineNumber: position.lineNumber,
          endLineNumber: position.lineNumber,
          startColumn: word.startColumn,
          endColumn: word.endColumn,
        }
        const line = model.getLineContent(position.lineNumber).slice(0, position.column - 1)
        const dotMatch = /(\w+)\.\s*$/.exec(line)
        const suggestions: languages.CompletionItem[] = []

        if (dotMatch && activeConn) {
          const table = dotMatch[1]
          let cols = schemaCache.current.get(table)
          if (!cols) {
            try {
              const schema = await schemaApi.getTableSchema(activeConn, table)
              cols = schema.columns.map((c) => c.name)
              schemaCache.current.set(table, cols)
            } catch {
              cols = []
            }
          }
          for (const col of cols) {
            suggestions.push({
              label: col,
              kind: monaco.languages.CompletionItemKind.Field,
              insertText: col,
              range,
            })
          }
        } else {
          for (const kw of SQL_KEYWORDS) {
            suggestions.push({
              label: kw,
              kind: monaco.languages.CompletionItemKind.Keyword,
              insertText: kw,
              range,
            })
          }
          for (const name of tableNames) {
            suggestions.push({
              label: name,
              kind: monaco.languages.CompletionItemKind.Class,
              insertText: name,
              range,
            })
          }
        }
        return { suggestions }
      },
    })
  }

  const formatEditorSQL = () => {
    try {
      const dialect = connectionTypeToSqlDialect(activeConnItem?.type)
      const formatted = formatSql(sql, dialect)
      setSql(formatted)
      editorRef.current?.setValue(formatted)
    } catch {
      pushToast(t('sql.formatError'), 'error')
    }
  }

  const resultColumns = useMemo(() => result?.columns?.map((c) => c.name) ?? [], [result])

  const copyCsv = () => {
    if (!result?.columns || !result.rows) return
    const format = loadCSVFormatPreference()
    navigator.clipboard.writeText(rowsToCSV(resultColumns, result.rows, format))
  }

  const openQueryExport = () => {
    if (!result?.columns || !result.rows || !activeConn) return
    openExport({
      source: 'query-result',
      connectionId: activeConn,
      availableColumns: resultColumns,
      rows: result.rows,
    })
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
        <Button onClick={() => void handleRun()} disabled={!activeConn || execute.isPending}>
          {t('sql.run')} (⌘↵)
        </Button>
        <Button variant="outline" size="sm" onClick={formatEditorSQL}>
          {t('sql.format')}
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
          height={`${editorHeight}px`}
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
              <>
                <Button size="sm" variant="outline" onClick={copyCsv}>
                  {t('sql.copyCsv')}
                </Button>
                <Button size="sm" variant="outline" onClick={openQueryExport}>
                  {t('csv.export')}
                </Button>
              </>
            )}
          </div>
          {result.kind === 'result' && isExplainResult(sql, result) && <ExplainTree result={result} />}
          {result.kind === 'result' && result.columns && !isExplainResult(sql, result) && (
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
                        {row[c.name] === null || row[c.name] === undefined ? (
                          <NullCell label={t('data.null')} />
                        ) : (
                          formatCell(row[c.name])
                        )}
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
