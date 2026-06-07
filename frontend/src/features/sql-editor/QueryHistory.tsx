import { useTranslation } from 'react-i18next'

export function QueryHistory({
  history,
  onSelect,
}: {
  history: string[]
  onSelect: (sql: string) => void
}) {
  const { t } = useTranslation()
  const items = history ?? []
  if (items.length === 0) return null
  return (
    <select
      className="rounded border border-border bg-transparent px-2 py-1 text-sm"
      onChange={(e) => e.target.value && onSelect(e.target.value)}
      defaultValue=""
    >
      <option value="">{t('sql.history')}</option>
      {items.map((h) => (
        <option key={h} value={h}>
          {h.slice(0, 60)}
        </option>
      ))}
    </select>
  )
}
