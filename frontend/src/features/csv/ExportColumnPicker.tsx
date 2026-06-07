import { useTranslation } from 'react-i18next'

export function ExportColumnPicker({
  columns,
  selected,
  onChange,
  disabled = false,
}: {
  columns: string[]
  selected: Set<string>
  onChange: (next: Set<string>) => void
  disabled?: boolean
}) {
  const { t } = useTranslation()
  const allSelected = columns.length > 0 && selected.size === columns.length

  const toggleAll = () => {
    onChange(allSelected ? new Set() : new Set(columns))
  }

  const toggleOne = (name: string) => {
    const next = new Set(selected)
    if (next.has(name)) next.delete(name)
    else next.add(name)
    onChange(next)
  }

  if (columns.length === 0) return null

  return (
    <div className="mb-3 space-y-2">
      <div className="flex items-center justify-between gap-2">
        <span className="text-sm font-medium">{t('csv.exportColumns')}</span>
        <button
          type="button"
          className="text-xs text-blue-600 hover:underline disabled:opacity-50 dark:text-blue-400"
          disabled={disabled}
          onClick={toggleAll}
        >
          {allSelected ? t('csv.deselectAllColumns') : t('csv.selectAllColumns')}
        </button>
      </div>
      <div className="max-h-40 space-y-1 overflow-y-auto rounded border border-border p-2">
        {columns.map((col) => (
          <label key={col} className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              disabled={disabled}
              checked={selected.has(col)}
              onChange={() => toggleOne(col)}
            />
            <span className="truncate font-mono text-xs">{col}</span>
          </label>
        ))}
      </div>
      <p className="text-xs text-muted">
        {t('csv.exportColumnsHint', { count: selected.size, total: columns.length })}
      </p>
    </div>
  )
}
