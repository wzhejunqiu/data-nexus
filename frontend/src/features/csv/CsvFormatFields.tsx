import { useTranslation } from 'react-i18next'
import type { CSVFormatOptions } from '@/lib/types/csv'

type Mode = 'export' | 'import'

export function CsvFormatFields({
  format,
  onChange,
  mode = 'export',
  disabled = false,
}: {
  format: CSVFormatOptions
  onChange: (next: CSVFormatOptions) => void
  mode?: Mode
  disabled?: boolean
}) {
  const { t } = useTranslation()
  const set = (patch: Partial<CSVFormatOptions>) => onChange({ ...format, ...patch })

  return (
    <div className="space-y-3 text-sm">
      <label className="flex flex-col gap-1">
        {t('csv.delimiter')}
        <select
          className="rounded border border-border bg-transparent px-2 py-1"
          value={format.delimiter}
          disabled={disabled}
          onChange={(e) => set({ delimiter: e.target.value })}
        >
          <option value=",">{t('csv.delimiterComma')}</option>
          <option value="\t">{t('csv.delimiterTab')}</option>
          <option value=";">{t('csv.delimiterSemicolon')}</option>
          <option value="|">{t('csv.delimiterPipe')}</option>
        </select>
      </label>
      <label className="flex flex-col gap-1">
        {t('csv.quoteChar')}
        <select
          className="rounded border border-border bg-transparent px-2 py-1"
          value={format.quoteChar}
          disabled={disabled}
          onChange={(e) => set({ quoteChar: e.target.value })}
        >
          <option value='"'>"</option>
          <option value="'">'</option>
        </select>
      </label>
      <label className="flex items-center gap-2">
        <input
          type="checkbox"
          disabled={disabled}
          checked={format.hasHeader}
          onChange={(e) => set({ hasHeader: e.target.checked })}
        />
        {t('csv.hasHeader')}
      </label>
      {mode === 'export' && (
        <>
          <label className="flex flex-col gap-1">
            {t('csv.lineEnding')}
            <select
              className="rounded border border-border bg-transparent px-2 py-1"
              value={format.lineEnding}
              disabled={disabled}
              onChange={(e) => set({ lineEnding: e.target.value })}
            >
              <option value="crlf">CRLF</option>
              <option value="lf">LF</option>
            </select>
          </label>
          <label className="flex flex-col gap-1">
            {t('csv.encoding')}
            <select
              className="rounded border border-border bg-transparent px-2 py-1"
              value={format.encoding}
              disabled={disabled}
              onChange={(e) => set({ encoding: e.target.value })}
            >
              <option value="utf-8">UTF-8</option>
              <option value="utf-8-bom">UTF-8 BOM</option>
            </select>
          </label>
        </>
      )}
      {mode === 'import' && (
        <>
          <label className="flex flex-col gap-1">
            {t('csv.encoding')}
            <select
              className="rounded border border-border bg-transparent px-2 py-1"
              value={format.encoding}
              disabled={disabled}
              onChange={(e) => set({ encoding: e.target.value })}
            >
              <option value="utf-8">UTF-8</option>
              <option value="utf-8-bom">UTF-8 BOM</option>
            </select>
          </label>
          <label className="flex flex-col gap-1">
            {t('csv.commentChar')}
            <input
              className="rounded border border-border bg-transparent px-2 py-1"
              value={format.commentChar ?? ''}
              maxLength={1}
              onChange={(e) => set({ commentChar: e.target.value })}
              placeholder="#"
            />
          </label>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={format.lazyQuotes ?? false}
              onChange={(e) => set({ lazyQuotes: e.target.checked })}
            />
            {t('csv.lazyQuotes')}
          </label>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={format.trimLeadingSpace ?? false}
              onChange={(e) => set({ trimLeadingSpace: e.target.checked })}
            />
            {t('csv.trimLeadingSpace')}
          </label>
        </>
      )}
    </div>
  )
}
