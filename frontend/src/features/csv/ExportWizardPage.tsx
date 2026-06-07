import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { ProgressBar } from '@/components/ui/ProgressBar'
import { CsvFormatFields } from '@/features/csv/CsvFormatFields'
import { ExportColumnPicker } from '@/features/csv/ExportColumnPicker'
import {
  getExportSteps,
  getExportWizardSteps,
  isInteractiveExportStep,
  isTableExport,
  stepLabelKey,
  type ExportStep,
} from '@/features/csv/exportWizardSteps'
import { runExport } from '@/features/csv/useExportRunner'
import { useExportProgress } from '@/features/csv/useExportProgress'
import { useToastStore } from '@/components/ui/Toast'
import { exportApi } from '@/lib/api/export'
import { dialogApi } from '@/lib/api/dialog'
import { formatError, isDialogCancelled, mapWailsError } from '@/lib/api/errors'
import type { CSVFormatOptions } from '@/lib/types/csv'
import { CSV_EXPORT_WARN_ROWS, loadCSVFormatPreference } from '@/lib/types/csv'
import type { ExportResult, ExportSource } from '@/stores/exportStore'
import { useExportStore } from '@/stores/exportStore'

export function ExportWizardPage() {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const session = useExportStore((s) => s.session)!
  const closeExport = useExportStore((s) => s.closeExport)

  const [source, setSource] = useState<ExportSource>(() => {
    if (
      session.source === 'table-page' &&
      session.totalRows != null &&
      session.rows != null &&
      session.totalRows > session.rows.length
    ) {
      return 'table-all'
    }
    return session.source
  })
  const [step, setStep] = useState<ExportStep>('options')
  const [format, setFormat] = useState<CSVFormatOptions>(() => loadCSVFormatPreference())
  const [selectedColumns, setSelectedColumns] = useState<Set<string>>(
    () => new Set(session.availableColumns),
  )
  const [filePath, setFilePath] = useState('')
  const [exportId, setExportId] = useState<string | null>(null)
  const [result, setResult] = useState<ExportResult | null>(null)
  const [running, setRunning] = useState(false)
  const progress = useExportProgress(exportId)
  const runStarted = useRef(false)
  const exportedCountRef = useRef(0)

  const steps = getExportSteps()
  const wizardSteps = getExportWizardSteps()
  const stepIndex = steps.indexOf(step)
  const wizardStepIndex = wizardSteps.indexOf(step)
  const orderedSelected = session.availableColumns.filter((c) => selectedColumns.has(c))
  const canProceedOptions = orderedSelected.length > 0

  const showLargeWarning =
    isTableExport(source) &&
    source === 'table-all' &&
    session.totalRows != null &&
    session.totalRows > CSV_EXPORT_WARN_ROWS

  const defaultFileName =
    source === 'query-result' ? 'query-result.csv' : `${session.tableName ?? 'export'}.csv`

  useEffect(() => {
    if (progress?.exported != null) {
      exportedCountRef.current = progress.exported
    }
  }, [progress?.exported])

  const goNext = useCallback(() => {
    const idx = steps.indexOf(step)
    if (idx < steps.length - 1) setStep(steps[idx + 1])
  }, [step, steps])

  const goBack = useCallback(() => {
    const idx = steps.indexOf(step)
    if (idx > 0) setStep(steps[idx - 1])
  }, [step, steps])

  const pickFile = async () => {
    try {
      const path = await dialogApi.saveFile(defaultFileName, [
        { displayName: 'CSV', pattern: '*.csv' },
      ])
      setFilePath(path)
    } catch (err) {
      const appErr = mapWailsError(err)
      if (isDialogCancelled(appErr)) return
      pushToast(formatError(t, appErr), 'error')
    }
  }

  const startExport = () => {
    const id = source === 'table-all' ? crypto.randomUUID() : null
    setExportId(id)
    setResult(null)
    runStarted.current = false
    setStep('progress')
  }

  const executeExport = useCallback(async () => {
    if (!filePath || runStarted.current) return
    runStarted.current = true
    setRunning(true)
    const exportResult = await runExport({
      source,
      connectionId: session.connectionId,
      tableName: session.tableName,
      columns: orderedSelected,
      format,
      filePath,
      rows: session.rows,
      exportId,
      t,
    })
    if (exportResult.status === 'success' && source === 'table-all') {
      exportResult.rowsExported = exportedCountRef.current
    }
    setResult(exportResult)
    setRunning(false)
    setStep('result')
  }, [
    filePath,
    source,
    session.connectionId,
    session.tableName,
    session.rows,
    orderedSelected,
    format,
    exportId,
    t,
  ])

  useEffect(() => {
    if (step === 'progress' && filePath && !result) {
      void executeExport()
    }
  }, [step, filePath, result, executeExport])

  const handleCancelProgress = async () => {
    if (exportId) {
      await exportApi.cancelExportTableCSV(exportId)
    }
  }

  const handleClose = () => {
    closeExport()
  }

  const handleBackFromError = () => {
    setResult(null)
    runStarted.current = false
    setRunning(false)
    setExportId(null)
    setStep(wizardSteps[wizardSteps.length - 1]!)
  }

  const footer = (() => {
    if (step === 'result') {
      return (
        <>
          {result?.status === 'error' && (
            <Button variant="outline" onClick={handleBackFromError}>
              {t('csv.back')}
            </Button>
          )}
          <Button onClick={handleClose}>{t('csv.closeWizard')}</Button>
        </>
      )
    }
    if (step === 'progress') {
      return (
        <>
          {source === 'table-all' && exportId && (
            <Button
              variant="outline"
              onClick={() => void handleCancelProgress()}
              disabled={!running}
            >
              {t('common.cancel')}
            </Button>
          )}
        </>
      )
    }
    return (
      <>
        {stepIndex > 0 && (
          <Button variant="outline" onClick={goBack}>
            {t('csv.back')}
          </Button>
        )}
        <Button variant="outline" onClick={handleClose}>
          {t('common.cancel')}
        </Button>
        {step === 'options' && (
          <Button onClick={goNext} disabled={!canProceedOptions}>
            {t('csv.next')}
          </Button>
        )}
        {step === 'destination' && (
          <>
            <Button variant="outline" onClick={() => void pickFile()}>
              {t('csv.choosePath')}
            </Button>
            <Button onClick={startExport} disabled={!filePath}>
              {t('csv.startExport')}
            </Button>
          </>
        )}
      </>
    )
  })()

  return (
    <div className="flex min-h-0 flex-1 flex-col bg-background">
      <div className="border-b border-border px-6 py-4">
        <h1 className="text-lg font-semibold">{t('csv.exportWizardTitle')}</h1>
        <p className="mt-1 text-sm text-muted">
          {isInteractiveExportStep(step)
            ? t('csv.stepOf', {
                current: wizardStepIndex + 1,
                total: wizardSteps.length,
                label: t(stepLabelKey(step)),
              })
            : t(stepLabelKey(step))}
        </p>
      </div>

      <div className="flex-1 overflow-auto px-6 py-6">
        <div className="mx-auto max-w-lg">
          {step === 'options' && (
            <div className="space-y-6">
              {isTableExport(source) && (
                <div className="space-y-4">
                  <p className="text-sm text-muted">{t('csv.scopeHint')}</p>
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="radio"
                      checked={source === 'table-page'}
                      onChange={() => setSource('table-page')}
                    />
                    {t('csv.exportPage')}
                  </label>
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="radio"
                      checked={source === 'table-all'}
                      onChange={() => setSource('table-all')}
                    />
                    {t('csv.exportAll')}
                  </label>
                  {showLargeWarning && (
                    <p className="text-sm text-amber-600 dark:text-amber-400">
                      {t('csv.exportLargeWarning', { total: session.totalRows })}
                    </p>
                  )}
                </div>
              )}
              <ExportColumnPicker
                columns={session.availableColumns}
                selected={selectedColumns}
                onChange={setSelectedColumns}
              />
              <CsvFormatFields format={format} onChange={setFormat} mode="export" />
            </div>
          )}

          {step === 'destination' && (
            <div className="space-y-4 text-sm">
              <p className="text-muted">{t('csv.destinationHint')}</p>
              <div className="rounded border border-border p-3 space-y-2">
                <SummaryRow label={t('csv.summaryScope')} value={scopeLabel(t, source)} />
                <SummaryRow
                  label={t('csv.summaryColumns')}
                  value={String(orderedSelected.length)}
                />
                <SummaryRow label={t('csv.filePath')} value={filePath || '—'} />
              </div>
            </div>
          )}

          {step === 'progress' && (
            <div className="space-y-4">
              {source === 'table-all' ? (
                <>
                  <ProgressBar indeterminate />
                  <p className="text-sm text-muted">
                    {t('csv.exportProgressCount', { exported: progress?.exported ?? 0 })}
                  </p>
                </>
              ) : (
                <p className="text-sm text-muted">{t('csv.exporting')}</p>
              )}
            </div>
          )}

          {step === 'result' && result && (
            <div className="space-y-4">
              <ResultBanner result={result} t={t} />
              {result.status === 'success' && (
                <dl className="space-y-2 rounded border border-border p-4 text-sm">
                  <SummaryRow label={t('csv.filePath')} value={result.filePath ?? '—'} />
                  {result.rowsExported != null && (
                    <SummaryRow label={t('csv.rowsExported')} value={String(result.rowsExported)} />
                  )}
                  <SummaryRow
                    label={t('csv.summaryColumns')}
                    value={String(result.columnCount ?? orderedSelected.length)}
                  />
                  {result.durationMs != null && (
                    <SummaryRow
                      label={t('csv.duration')}
                      value={t('csv.durationMs', { ms: result.durationMs })}
                    />
                  )}
                </dl>
              )}
              {result.message && result.status !== 'success' && (
                <p className="text-sm text-muted">{result.message}</p>
              )}
            </div>
          )}
        </div>
      </div>

      <div className="flex justify-end gap-2 border-t border-border px-6 py-4">{footer}</div>
    </div>
  )
}

function SummaryRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid grid-cols-[8rem_1fr] items-baseline gap-x-4">
      <dt className="text-right text-muted">{label}</dt>
      <dd className="min-w-0 break-all font-mono text-xs">{value}</dd>
    </div>
  )
}

function scopeLabel(t: (key: string) => string, source: ExportSource): string {
  if (source === 'table-page') return t('csv.exportPage')
  if (source === 'table-all') return t('csv.exportAll')
  return t('csv.queryResult')
}

function ResultBanner({ result, t }: { result: ExportResult; t: (key: string) => string }) {
  const cls =
    result.status === 'success'
      ? 'border-green-600/40 bg-green-50 text-green-800 dark:bg-green-950/30 dark:text-green-300'
      : result.status === 'cancelled'
        ? 'border-amber-600/40 bg-amber-50 text-amber-800 dark:bg-amber-950/30 dark:text-amber-300'
        : 'border-red-600/40 bg-red-50 text-red-800 dark:bg-red-950/30 dark:text-red-300'
  const msg =
    result.status === 'success'
      ? t('csv.exportResultSuccess')
      : result.status === 'cancelled'
        ? t('csv.exportResultCancelled')
        : t('csv.exportResultError')
  return <div className={`rounded border px-4 py-3 text-sm font-medium ${cls}`}>{msg}</div>
}
