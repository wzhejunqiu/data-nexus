import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { filtersToSQL } from '@/lib/filtersToSQL'
import type { ColumnInfo, RowFilter, FilterOperator } from '@/lib/types'
import { useWorkspaceStore } from '@/stores/workspaceStore'

const OPERATORS: FilterOperator[] = [
  'eq',
  'ne',
  'gt',
  'gte',
  'lt',
  'lte',
  'like',
  'is_null',
  'is_not_null',
  'in',
]

const EMPTY_FILTER = (): RowFilter => ({ column: '', operator: 'eq', value: '' })

export function FilterBuilder({
  tableName,
  columns,
  filters,
  sort,
  order,
  onChange,
  onApply,
}: {
  tableName: string
  columns: ColumnInfo[]
  filters: RowFilter[]
  sort: string
  order: 'asc' | 'desc'
  onChange: (filters: RowFilter[]) => void
  onApply: () => void
}) {
  const { t } = useTranslation()
  const setActiveTab = useWorkspaceStore((s) => s.setActiveTab)
  const setPendingSql = useWorkspaceStore((s) => s.setPendingSql)

  const rows = filters.length > 0 ? filters : [EMPTY_FILTER()]

  const updateRow = (idx: number, patch: Partial<RowFilter>) => {
    const next = rows.map((r, i) => (i === idx ? { ...r, ...patch } : r))
    onChange(next)
  }

  const addRow = () => onChange([...rows, EMPTY_FILTER()])
  const removeRow = (idx: number) => onChange(rows.filter((_, i) => i !== idx))

  const activeFilters = rows.filter((r) => r.column && r.operator)

  const openInSqlEditor = () => {
    const sql = filtersToSQL(tableName, activeFilters, sort || undefined, order)
    setPendingSql(sql)
    setActiveTab('sql')
  }

  return (
    <div className="mb-3 space-y-2 rounded border border-border p-3">
      <p className="text-xs font-semibold text-muted">{t('filter.title')}</p>
      {rows.map((row, idx) => (
        <div key={idx} className="flex flex-wrap items-center gap-2">
          <select
            className="rounded border border-border bg-transparent px-2 py-1 text-xs"
            value={row.column}
            onChange={(e) => updateRow(idx, { column: e.target.value })}
          >
            <option value="">{t('filter.selectColumn')}</option>
            {columns.map((c) => (
              <option key={c.name} value={c.name}>
                {c.name}
              </option>
            ))}
          </select>
          <select
            className="rounded border border-border bg-transparent px-2 py-1 text-xs"
            value={row.operator}
            onChange={(e) => updateRow(idx, { operator: e.target.value as FilterOperator })}
          >
            {OPERATORS.map((op) => (
              <option key={op} value={op}>
                {t(`filter.op.${op}`)}
              </option>
            ))}
          </select>
          {!['is_null', 'is_not_null'].includes(row.operator) && (
            <input
              className="min-w-[120px] flex-1 rounded border border-border bg-transparent px-2 py-1 text-xs"
              placeholder={row.operator === 'in' ? t('filter.inPlaceholder') : t('filter.value')}
              value={row.operator === 'in' ? (row.values ?? []).join(', ') : (row.value ?? '')}
              onChange={(e) => {
                if (row.operator === 'in') {
                  updateRow(idx, {
                    values: e.target.value
                      .split(',')
                      .map((s) => s.trim())
                      .filter(Boolean),
                  })
                } else {
                  updateRow(idx, { value: e.target.value })
                }
              }}
            />
          )}
          <Button size="sm" variant="outline" onClick={() => removeRow(idx)}>
            {t('filter.remove')}
          </Button>
        </div>
      ))}
      <div className="flex flex-wrap gap-2">
        <Button size="sm" variant="outline" onClick={addRow}>
          {t('filter.add')}
        </Button>
        <Button size="sm" onClick={onApply}>
          {t('filter.apply')}
        </Button>
        <Button size="sm" variant="outline" onClick={() => onChange([])}>
          {t('filter.clear')}
        </Button>
        <Button
          size="sm"
          variant="outline"
          onClick={openInSqlEditor}
          disabled={activeFilters.length === 0}
        >
          {t('filter.editInSql')}
        </Button>
      </div>
    </div>
  )
}
