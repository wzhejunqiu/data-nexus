import Editor, { type OnMount } from '@monaco-editor/react'
import { useMemo, useState } from 'react'
import type { SqlDialect } from '@/lib/sql/dialect'
import { formatSqlSafe } from '@/lib/sql/format'
import { resolveTheme, useThemeStore } from '@/stores/themeStore'

const EDITOR_LINE_HEIGHT = 19
const MIN_EDITOR_LINES = 3
const DEFAULT_MAX_LINES = 30

export function SqlCodeView({
  sql,
  dialect = 'sql',
  maxLines = DEFAULT_MAX_LINES,
}: {
  sql: string
  dialect?: SqlDialect
  maxLines?: number
}) {
  const themeMode = useThemeStore((s) => s.mode)
  const monacoTheme = resolveTheme(themeMode) === 'dark' ? 'vs-dark' : 'vs'
  const formatted = useMemo(() => formatSqlSafe(sql, dialect), [sql, dialect])
  const [editorHeight, setEditorHeight] = useState(MIN_EDITOR_LINES * EDITOR_LINE_HEIGHT)

  const handleMount: OnMount = (editor) => {
    const updateHeight = () => {
      const lines = Math.max(MIN_EDITOR_LINES, editor.getContentHeight() / EDITOR_LINE_HEIGHT)
      setEditorHeight(Math.min(maxLines, lines) * EDITOR_LINE_HEIGHT)
    }
    updateHeight()
    editor.onDidContentSizeChange(updateHeight)
  }

  return (
    <div className="overflow-hidden rounded border border-border" data-testid="sql-code-view">
      <Editor
        height={`${editorHeight}px`}
        defaultLanguage="sql"
        theme={monacoTheme}
        value={formatted}
        onMount={handleMount}
        options={{
          readOnly: true,
          minimap: { enabled: false },
          fontSize: 13,
          lineNumbers: 'on',
          scrollBeyondLastLine: false,
          wordWrap: 'on',
          automaticLayout: true,
        }}
      />
    </div>
  )
}
