import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { ProgressBar } from '@/components/ui/ProgressBar'
import { CsvFormatFields } from '@/features/csv/CsvFormatFields'
import { useExportProgress } from '@/features/csv/useExportProgress'
import type { CSVFormatOptions } from '@/lib/types/csv'
import { CSV_EXPORT_WARN_ROWS, loadCSVFormatPreference, saveCSVFormatPreference } from '@/lib/types/csv'

export function ExportCsvDialog({
  open,
  title,
  mode,
  totalRows,
  onOpenChange,
  onExport,
  onCancelExport,
}: {
  open: boolean
  title: string
  mode: 'page' | 'all'
  totalRows?: number
  onOpenChange: (open: boolean) => void
  onExport: (format: CSVFormatOptions, exportId: string | null) => void | Promise<void>
  onCancelExport?: (exportId: string) => void | Promise<void>
}) {
  const { t } = useTranslation()
  const [format, setFormat] = useState<CSVFormatOptions>(() => loadCSVFormatPreference())
  const [busy, setBusy] = useState(false)
  const [exportId, setExportId] = useState<string | null>(null)
  const progress = useExportProgress(exportId)

  const showLargeWarning =
    mode === 'all' && totalRows != null && totalRows > CSV_EXPORT_WARN_ROWS

  const handleExport = async () => {
    const id = mode === 'all' ? crypto.randomUUID() : null
    setExportId(id)
    setBusy(true)
    try {
      saveCSVFormatPreference(format)
      await onExport(format, id)
      onOpenChange(false)
    } finally {
      setBusy(false)
      setExportId(null)
    }
  }

  const handleCancel = () => {
    if (busy && exportId && onCancelExport) {
      void onCancelExport(exportId)
      return
    }
    onOpenChange(false)
  }

  const handleOpenChange = (next: boolean) => {
    if (!next && busy) {
      handleCancel()
      return
    }
    onOpenChange(next)
  }

  return (
    <Dialog
      open={open}
      onOpenChange={handleOpenChange}
      title={title}
      footer={
        <>
          <Button variant="outline" onClick={handleCancel} disabled={false}>
            {t('common.cancel')}
          </Button>
          <Button onClick={() => void handleExport()} disabled={busy}>
            {busy ? t('csv.exporting') : t('csv.export')}
          </Button>
        </>
      }
    >
      {showLargeWarning && (
        <p className="mb-3 text-sm text-amber-600 dark:text-amber-400">
          {t('csv.exportLargeWarning', { total: totalRows })}
        </p>
      )}
      {busy && mode === 'all' && (
        <div className="mb-3 space-y-1">
          <ProgressBar indeterminate />
          <p className="text-xs text-muted">
            {t('csv.exportProgressCount', { exported: progress?.exported ?? 0 })}
          </p>
        </div>
      )}
      <CsvFormatFields format={format} onChange={setFormat} mode="export" disabled={busy} />
    </Dialog>
  )
}
