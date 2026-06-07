import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { useToastStore } from '@/components/ui/Toast'
import { CsvFormatFields } from '@/features/csv/CsvFormatFields'
import { SqlCodeView } from '@/features/sql-editor/SqlCodeView'
import {
  getImportSteps,
  getImportWizardSteps,
  isInteractiveImportStep,
  stepLabelKey,
  type ImportStep,
} from '@/features/csv/importWizardSteps'
import {
  buildCreateTableSQL,
  buildImportColumnMap,
  buildNewTableColumnSpecs,
  SQL_COLUMN_TYPES,
  type NewTableColumnSpec,
} from '@/lib/csvImport'
import { connectionApi } from '@/lib/api/connection'
import { dialogApi } from '@/lib/api/dialog'
import { formatError, isDialogCancelled, mapWailsError } from '@/lib/api/errors'
import { importApi } from '@/lib/api/import'
import { connectionTypeToSqlDialect } from '@/lib/sql/dialect'
import { schemaApi } from '@/lib/api/schema'
import type { CSVFormatOptions, CSVPreview } from '@/lib/types/csv'
import { loadCSVFormatPreference, saveCSVFormatPreference } from '@/lib/types/csv'
import type { ImportResult } from '@/stores/importStore'
import { useImportStore } from '@/stores/importStore'

export function ImportWizardPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const pushToast = useToastStore((s) => s.push)
  const session = useImportStore((s) => s.session)!
  const closeImport = useImportStore((s) => s.closeImport)

  const [step, setStep] = useState<ImportStep>('source')
  const [filePath, setFilePath] = useState('')
  const [format, setFormat] = useState<CSVFormatOptions>(() => loadCSVFormatPreference())
  const [preview, setPreview] = useState<CSVPreview | null>(null)
  const [targetMode, setTargetMode] = useState<'new' | 'existing'>('existing')
  const [newTableName, setNewTableName] = useState('')
  const [targetTable, setTargetTable] = useState(session.defaultTable ?? '')
  const [importMode, setImportMode] = useState<'append' | 'update'>('append')
  const [columnMap, setColumnMap] = useState<Record<string, string>>({})
  const [newColumnSpecs, setNewColumnSpecs] = useState<Record<string, NewTableColumnSpec>>({})
  const [result, setResult] = useState<ImportResult | null>(null)
  const runStarted = useRef(false)

  const steps = getImportSteps()
  const wizardSteps = getImportWizardSteps()
  const stepIndex = steps.indexOf(step)
  const wizardStepIndex = wizardSteps.indexOf(step)

  const createTableSQL = useMemo(() => {
    if (targetMode !== 'new' || !preview) return null
    return buildCreateTableSQL(
      newTableName.trim(),
      preview.headers,
      columnMap,
      newColumnSpecs,
    )
  }, [targetMode, preview, newTableName, columnMap, newColumnSpecs])

  const { data: connections } = useQuery({
    queryKey: ['connections'],
    queryFn: () => connectionApi.list(),
  })
  const connItem = connections?.items.find((c) => c.id === session.connectionId)
  const sqlDialect = connectionTypeToSqlDialect(connItem?.type)

  const { data: tables } = useQuery({
    queryKey: ['schema', 'tables', session.connectionId],
    queryFn: () => schemaApi.listTables(session.connectionId),
    enabled: step !== 'source',
  })

  const { data: targetSchema } = useQuery({
    queryKey: ['schema', session.connectionId, targetTable],
    queryFn: () => schemaApi.getTableSchema(session.connectionId, targetTable),
    enabled: targetMode === 'existing' && !!targetTable && step !== 'source',
  })

  const targetHasPK = targetSchema?.columns.some((c) => c.primaryKey) ?? false
  const updateBlocked =
    targetMode === 'existing' && importMode === 'update' && !!targetTable && !targetHasPK

  const loadPreview = useMutation({
    mutationFn: () =>
      importApi.parseCSVPreview({ filePath, maxRows: 20, format }),
    onSuccess: (res) => {
      setPreview(res)
      setStep('target')
    },
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

  const goBack = useCallback(() => {
    const idx = steps.indexOf(step)
    if (idx > 0) setStep(steps[idx - 1])
  }, [step, steps])

  const hasMappedColumns = Object.values(columnMap).some((c) => c.trim() !== '')

  const goToMapping = () => {
    if (!preview) return
    if (targetMode === 'existing') {
      if (!targetTable) {
        pushToast(t('csv.selectTargetTable'), 'error')
        return
      }
      const cols = targetSchema?.columns.map((c) => c.name) ?? []
      setColumnMap(buildImportColumnMap(preview.headers, cols))
    } else {
      if (!newTableName.trim()) {
        pushToast(t('csv.newTableNameRequired'), 'error')
        return
      }
      setColumnMap(buildImportColumnMap(preview.headers))
      setNewColumnSpecs(buildNewTableColumnSpecs(preview.headers, preview.rows))
    }
    setStep('mapping')
  }

  const targetColumns = targetSchema?.columns.map((c) => c.name) ?? []

  const pickFile = async () => {
    try {
      const path = await dialogApi.openCSVFile()
      setFilePath(path)
    } catch (err) {
      const appErr = mapWailsError(err)
      if (isDialogCancelled(appErr)) return
      pushToast(formatError(t, appErr), 'error')
    }
  }

  const startImport = () => {
    if (!hasMappedColumns) {
      pushToast(t('csv.noMappedColumns'), 'error')
      return
    }
    saveCSVFormatPreference(format)
    setResult(null)
    runStarted.current = false
    setStep('progress')
  }

  const executeImport = useCallback(async () => {
    if (runStarted.current) return
    runStarted.current = true
    const start = performance.now()
    try {
      const res = await importApi.importCSV({
        connectionId: session.connectionId,
        targetTable: targetMode === 'existing' ? targetTable : '',
        newTableName: targetMode === 'new' ? newTableName : '',
        mode: importMode,
        columnMap,
        newTableColumns: targetMode === 'new' ? newColumnSpecs : undefined,
        filePath,
        upsertKeys: [],
        format,
      })
      setResult({
        status: 'success',
        rowsInserted: res.rowsInserted,
        rowsUpdated: res.rowsUpdated,
        durationMs: Math.round(performance.now() - start),
      })
    } catch (err) {
      setResult({
        status: 'error',
        message: formatError(t, err),
        durationMs: Math.round(performance.now() - start),
      })
    } finally {
      setStep('result')
    }
  }, [
    session.connectionId,
    targetMode,
    targetTable,
    newTableName,
    importMode,
    columnMap,
    newColumnSpecs,
    filePath,
    format,
    t,
  ])

  useEffect(() => {
    if (step === 'progress' && !result) {
      void executeImport()
    }
  }, [step, result, executeImport])

  const handleClose = () => {
    if (result?.status === 'success') {
      void qc.invalidateQueries({ queryKey: ['schema', session.connectionId] })
      void qc.invalidateQueries({ queryKey: ['tables', session.connectionId] })
      void qc.invalidateQueries({ queryKey: ['rows', session.connectionId] })
    }
    closeImport()
  }

  const handleBackFromError = () => {
    setResult(null)
    runStarted.current = false
    setStep(wizardSteps[wizardSteps.length - 1]!)
  }

  const resolvedTarget =
    targetMode === 'existing' ? targetTable : newTableName.trim()

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
      return null
    }
    return (
      <>
        {stepIndex > 0 && (
          <Button variant="outline" onClick={goBack}>
            {t('csv.back')}
          </Button>
        )}
        <Button variant="outline" onClick={closeImport}>
          {t('common.cancel')}
        </Button>
        {step === 'source' && !filePath && (
          <Button onClick={() => void pickFile()}>{t('csv.selectFile')}</Button>
        )}
        {step === 'source' && filePath && (
          <Button
            onClick={() => {
              saveCSVFormatPreference(format)
              loadPreview.mutate()
            }}
            disabled={loadPreview.isPending}
          >
            {t('csv.preview')}
          </Button>
        )}
        {step === 'target' && (
          <Button onClick={goToMapping} disabled={updateBlocked}>
            {t('csv.next')}
          </Button>
        )}
        {step === 'mapping' && (
          <Button
            onClick={startImport}
            disabled={updateBlocked || !hasMappedColumns}
          >
            {t('csv.startImport')}
          </Button>
        )}
      </>
    )
  })()

  const wideContent =
    step === 'mapping' && targetMode === 'new'

  return (
    <div className="flex min-h-0 flex-1 flex-col bg-background">
      <div className="border-b border-border px-6 py-4">
        <h1 className="text-lg font-semibold">{t('csv.importWizardTitle')}</h1>
        <p className="mt-1 text-sm text-muted">
          {isInteractiveImportStep(step)
            ? t('csv.stepOf', {
                current: wizardStepIndex + 1,
                total: wizardSteps.length,
                label: t(stepLabelKey(step)),
              })
            : t(stepLabelKey(step))}
        </p>
      </div>

      <div className="flex-1 overflow-auto px-6 py-6">
        <div className={`mx-auto w-full ${wideContent ? 'max-w-3xl' : 'max-w-lg'}`}>
          {step === 'source' && (
            <div className="space-y-3">
              {!filePath ? (
                <p className="text-sm text-muted">{t('csv.importHint')}</p>
              ) : (
                <>
                  <p className="text-xs text-muted">{filePath}</p>
                  <CsvFormatFields format={format} onChange={setFormat} mode="import" />
                </>
              )}
            </div>
          )}

          {step === 'target' && preview && (
            <div className="space-y-3 text-sm">
              <p className="text-muted">
                {t('csv.fileRowCount', { count: preview.rowCount })}
              </p>
              <div className="max-h-40 overflow-auto rounded border border-border">
                <PreviewTable preview={preview} />
              </div>
              <label className="flex items-center gap-2">
                <input
                  type="radio"
                  checked={targetMode === 'existing'}
                  onChange={() => setTargetMode('existing')}
                />
                {t('csv.existingTable')}
              </label>
              {targetMode === 'existing' && (
                <select
                  className="w-full rounded border border-border bg-transparent px-2 py-1"
                  value={targetTable}
                  onChange={(e) => setTargetTable(e.target.value)}
                >
                  <option value="">—</option>
                  {tables?.items
                    .filter((tbl) => tbl.type === 'table')
                    .map((tbl) => (
                      <option key={tbl.name} value={tbl.name}>
                        {tbl.name}
                      </option>
                    ))}
                </select>
              )}
              <label className="flex items-center gap-2">
                <input
                  type="radio"
                  checked={targetMode === 'new'}
                  onChange={() => setTargetMode('new')}
                />
                {t('csv.newTable')}
              </label>
              {targetMode === 'new' && (
                <input
                  className="w-full rounded border border-border bg-transparent px-2 py-1"
                  value={newTableName}
                  onChange={(e) => setNewTableName(e.target.value)}
                  placeholder={t('csv.newTableName')}
                />
              )}
              <label className="flex flex-col gap-1">
                {t('csv.importMode')}
                <select
                  className="rounded border border-border bg-transparent px-2 py-1"
                  value={importMode}
                  onChange={(e) => setImportMode(e.target.value as 'append' | 'update')}
                >
                  <option value="append">{t('csv.modeAppend')}</option>
                  <option
                    value="update"
                    disabled={targetMode === 'existing' && !!targetTable && !targetHasPK}
                  >
                    {t('csv.modeUpdate')}
                  </option>
                </select>
              </label>
              {updateBlocked && (
                <p className="text-sm text-amber-600 dark:text-amber-400">
                  {t('csv.updateRequiresPK')}
                </p>
              )}
            </div>
          )}

          {step === 'mapping' && preview && targetMode === 'existing' && (
            <div className="space-y-4 text-sm">
              <div className="space-y-2">
                {preview.headers.map((h) => (
                  <label key={h} className="flex items-center gap-2">
                    <span className="w-32 truncate font-mono text-xs">{h}</span>
                    <span>→</span>
                    <select
                      className="flex-1 rounded border border-border bg-transparent px-2 py-1 font-mono text-xs"
                      value={columnMap[h] ?? ''}
                      onChange={(e) => setColumnMap({ ...columnMap, [h]: e.target.value })}
                    >
                      <option value="">{t('csv.skipColumn')}</option>
                      {targetColumns.map((col) => (
                        <option key={col} value={col}>
                          {col}
                        </option>
                      ))}
                    </select>
                  </label>
                ))}
              </div>
              <ImportSummary
                filePath={filePath}
                resolvedTarget={resolvedTarget}
                importMode={importMode}
                columnMap={columnMap}
                fileRowCount={preview.rowCount}
                t={t}
              />
            </div>
          )}

          {step === 'mapping' && preview && targetMode === 'new' && (
            <div className="space-y-4 text-sm">
              <p className="text-muted">{t('csv.mappingNewTableHint')}</p>
              <div className="overflow-x-auto rounded border border-border">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="border-b border-border bg-muted/30">
                      <th className="px-2 py-2 text-left font-medium">CSV</th>
                      <th className="px-2 py-2 text-left font-medium">{t('csv.columnName')}</th>
                      <th className="px-2 py-2 text-left font-medium">{t('csv.columnType')}</th>
                      <th className="px-2 py-2 text-left font-medium">{t('csv.primaryKey')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {preview.headers.map((h) => {
                      const spec = newColumnSpecs[h]
                      if (!spec) return null
                      return (
                        <tr key={h} className="border-b border-border/40">
                          <td className="px-2 py-2 font-mono">{h}</td>
                          <td className="px-2 py-2">
                            <input
                              className="w-full min-w-[6rem] rounded border border-border bg-transparent px-2 py-1 font-mono"
                              value={columnMap[h] ?? ''}
                              onChange={(e) =>
                                setColumnMap({ ...columnMap, [h]: e.target.value })
                              }
                              placeholder={t('csv.skipColumn')}
                            />
                          </td>
                          <td className="px-2 py-2">
                            <select
                              className="w-full rounded border border-border bg-transparent px-2 py-1 font-mono"
                              value={spec.dataType}
                              disabled={!columnMap[h]}
                              onChange={(e) =>
                                setNewColumnSpecs({
                                  ...newColumnSpecs,
                                  [h]: {
                                    ...spec,
                                    dataType: e.target.value as NewTableColumnSpec['dataType'],
                                  },
                                })
                              }
                            >
                              {SQL_COLUMN_TYPES.map((typ) => (
                                <option key={typ} value={typ}>
                                  {typ}
                                </option>
                              ))}
                            </select>
                          </td>
                          <td className="px-2 py-2">
                            <input
                              type="checkbox"
                              checked={spec.primaryKey}
                              disabled={!columnMap[h]}
                              onChange={(e) =>
                                setNewColumnSpecs({
                                  ...newColumnSpecs,
                                  [h]: { ...spec, primaryKey: e.target.checked },
                                })
                              }
                            />
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
              {createTableSQL && (
                <div className="space-y-2">
                  <p className="text-muted">{t('csv.ddlPreviewHint')}</p>
                  <SqlCodeView sql={createTableSQL} dialect={sqlDialect} />
                </div>
              )}
              <ImportSummary
                filePath={filePath}
                resolvedTarget={resolvedTarget}
                importMode={importMode}
                columnMap={columnMap}
                fileRowCount={preview.rowCount}
                t={t}
              />
            </div>
          )}

          {step === 'progress' && (
            <p className="text-sm text-muted">{t('csv.importing')}</p>
          )}

          {step === 'result' && result && (
            <div className="space-y-4">
              <ResultBanner result={result} t={t} />
              {result.status === 'success' && (
                <dl className="space-y-2 rounded border border-border p-4 text-sm">
                  <SummaryRow
                    label={t('csv.rowsInserted')}
                    value={String(result.rowsInserted ?? 0)}
                  />
                  <SummaryRow
                    label={t('csv.rowsUpdated')}
                    value={String(result.rowsUpdated ?? 0)}
                  />
                  <SummaryRow label={t('csv.summaryTarget')} value={resolvedTarget || '—'} />
                  {result.durationMs != null && (
                    <SummaryRow
                      label={t('csv.duration')}
                      value={t('csv.durationMs', { ms: result.durationMs })}
                    />
                  )}
                </dl>
              )}
              {result.message && result.status === 'error' && (
                <p className="text-sm text-muted">{result.message}</p>
              )}
            </div>
          )}
        </div>
      </div>

      {footer && (
        <div className="flex justify-end gap-2 border-t border-border px-6 py-4">{footer}</div>
      )}
    </div>
  )
}

function ImportSummary({
  filePath,
  resolvedTarget,
  importMode,
  columnMap,
  fileRowCount,
  t,
}: {
  filePath: string
  resolvedTarget: string
  importMode: 'append' | 'update'
  columnMap: Record<string, string>
  fileRowCount?: number
  t: (key: string) => string
}) {
  return (
    <div className="rounded border border-border p-3 space-y-2">
      <p className="text-muted">{t('csv.confirmHint')}</p>
      <SummaryRow label={t('csv.filePath')} value={filePath} />
      {fileRowCount != null && (
        <SummaryRow
          label={t('csv.summaryFileRows')}
          value={String(fileRowCount)}
        />
      )}
      <SummaryRow label={t('csv.summaryTarget')} value={resolvedTarget || '—'} />
      <SummaryRow
        label={t('csv.importMode')}
        value={importMode === 'append' ? t('csv.modeAppend') : t('csv.modeUpdate')}
      />
      <SummaryRow
        label={t('csv.summaryColumns')}
        value={String(Object.values(columnMap).filter(Boolean).length)}
      />
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

function ResultBanner({
  result,
  t,
}: {
  result: ImportResult
  t: (key: string) => string
}) {
  const cls =
    result.status === 'success'
      ? 'border-green-600/40 bg-green-50 text-green-800 dark:bg-green-950/30 dark:text-green-300'
      : 'border-red-600/40 bg-red-50 text-red-800 dark:bg-red-950/30 dark:text-red-300'
  const msg =
    result.status === 'success' ? t('csv.importResultSuccess') : t('csv.importResultError')
  return (
    <div className={`rounded border px-4 py-3 text-sm font-medium ${cls}`}>{msg}</div>
  )
}

function PreviewTable({ preview }: { preview: CSVPreview }) {
  return (
    <table className="w-full text-xs">
      <thead>
        <tr>
          {preview.headers.map((h) => (
            <th key={h} className="border-b border-border px-2 py-1 text-left">
              {h}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {preview.rows.map((row, i) => (
          <tr key={i}>
            {preview.headers.map((h) => (
              <td key={h} className="border-b border-border/40 px-2 py-1 font-mono">
                {String(row[h] ?? '')}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  )
}
