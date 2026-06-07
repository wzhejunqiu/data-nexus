import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

export function FtsSearchBar({
  defaultValue,
  onChange,
}: {
  defaultValue: string
  onChange: (search: string) => void
}) {
  const { t } = useTranslation()
  const [draft, setDraft] = useState(defaultValue)

  useEffect(() => {
    const id = window.setTimeout(() => onChange(draft), 300)
    return () => window.clearTimeout(id)
  }, [draft, onChange])

  return (
    <label className="flex items-center gap-2 text-sm">
      {t('fts.search')}
      <input
        className="min-w-[200px] rounded border border-border bg-transparent px-2 py-1 text-xs"
        placeholder={t('fts.placeholder')}
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
      />
    </label>
  )
}
