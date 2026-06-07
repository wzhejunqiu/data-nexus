import { useMutation, useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { useToastStore } from '@/components/ui/Toast'
import { CsvFormatFields } from '@/features/csv/CsvFormatFields'
import { dialogApi } from '@/lib/api/dialog'
import { formatError, isDialogCancelled, mapWailsError } from '@/lib/api/errors'
import { importApi } from '@/lib/api/import'
import { schemaApi } from '@/lib/api/schema'
import type { CSVFormatOptions, CSVPreview } from '@/lib/types/csv'
import { loadCSVFormatPreference, saveCSVFormatPreference } from '@/lib/types/csv'
import { buildImportColumnMap } from '@/lib/csvImport'

type Step = 'file' | 'format' | 'target' | 'mapping' | 'confirm'

export function ImportWizard({
  connectionId,
  defaultTable,
  onClose,
  onDone,
}: {
  connectionId: string
  defaultTable?: string
  onClose: () => void
  onDone: () => void
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const [step, setStep] = useState<Step>('file')
  const [filePath, setFilePath] = useState('')
  const [format, setFormat] = useState<CSVFormatOptions>(() => loadCSVFormatPreference())
  const [preview, setPreview] = useState<CSVPreview | null>(null)
  const [targetMode, setTargetMode] = useState<'new' | 'existing'>('existing')
  const [newTableName, setNewTableName] = useState('')
  const [targetTable, setTargetTable] = useState(defaultTable ?? '')
  const [importMode, setImportMode] = useState<'append' | 'update'>('append')
  const [columnMap, setColumnMap] = useState<Record<string, string>>({})

  const { data: tables } = useQuery({
    queryKey: ['schema', 'tables', connectionId],
    queryFn: () => schemaApi.listTables(connectionId),
    enabled: step !== 'file',
  })

  const { data: targetSchema } = useQuery({
    queryKey: ['schema', connectionId, targetTable],
    queryFn: () => schemaApi.getTableSchema(connectionId, targetTable),
    enabled: targetMode === 'existing' && !!targetTable && step !== 'file',
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

  const runImport = useMutation({
    mutationFn: () =>
      importApi.importCSV({
        connectionId,
        targetTable: targetMode === 'existing' ? targetTable : '',
        newTableName: targetMode === 'new' ? newTableName : '',
        mode: importMode,
        columnMap,
        filePath,
        upsertKeys: [],
        format,
      }),
    onSuccess: (res) => {
      pushToast(
        t('csv.importSuccess', { inserted: res.rowsInserted, updated: res.rowsUpdated }),
        'info',
      )
      onDone()
    },
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

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
    }
    setStep('mapping')
  }

  const targetColumns = targetSchema?.columns.map((c) => c.name) ?? []

  const pickFile = async () => {
    try {
      const path = await dialogApi.openCSVFile()
      setFilePath(path)
      setStep('format')
    } catch (err) {
      const appErr = mapWailsError(err)
      if (isDialogCancelled(appErr)) return
      pushToast(formatError(t, appErr), 'error')
    }
  }

  return (
    <Dialog
      open
      onOpenChange={(open) => !open && onClose()}
      title={t('csv.importTitle')}
      footer={
        <>
          <Button variant="outline" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          {step === 'file' && (
            <Button onClick={() => void pickFile()}>{t('csv.selectFile')}</Button>
          )}
          {step === 'format' && (
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
            <Button onClick={() => setStep('confirm')} disabled={updateBlocked}>
              {t('csv.next')}
            </Button>
          )}
          {step === 'confirm' && (
            <Button onClick={() => runImport.mutate()} disabled={runImport.isPending || updateBlocked}>
              {t('csv.import')}
            </Button>
          )}
        </>
      }
    >
      {step === 'file' && <p className="text-sm text-muted">{t('csv.importHint')}</p>}
      {step === 'format' && (
        <div className="space-y-3">
          <p className="text-xs text-muted">{filePath}</p>
          <CsvFormatFields format={format} onChange={setFormat} mode="import" />
        </div>
      )}
      {step === 'target' && preview && (
        <div className="space-y-3 text-sm">
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
              {tables?.items.filter((t) => t.type === 'table').map((t) => (
                <option key={t.name} value={t.name}>
                  {t.name}
                </option>
              ))}
            </select>
          )}
          <label className="flex items-center gap-2">
            <input type="radio" checked={targetMode === 'new'} onChange={() => setTargetMode('new')} />
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
              <option value="update" disabled={targetMode === 'existing' && !!targetTable && !targetHasPK}>
                {t('csv.modeUpdate')}
              </option>
            </select>
          </label>
          {updateBlocked && (
            <p className="text-sm text-amber-600 dark:text-amber-400">{t('csv.updateRequiresPK')}</p>
          )}
        </div>
      )}
      {step === 'mapping' && preview && (
        <div className="space-y-2 text-sm">
          {preview.headers.map((h) => (
            <label key={h} className="flex items-center gap-2">
              <span className="w-32 truncate font-mono text-xs">{h}</span>
              <span>→</span>
              {targetMode === 'existing' ? (
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
              ) : (
                <input
                  className="flex-1 rounded border border-border bg-transparent px-2 py-1 font-mono text-xs"
                  value={columnMap[h] ?? ''}
                  onChange={(e) => setColumnMap({ ...columnMap, [h]: e.target.value })}
                />
              )}
            </label>
          ))}
        </div>
      )}
      {step === 'confirm' && (
        <p className="text-sm">{t('csv.confirmImport', { file: filePath, mode: importMode })}</p>
      )}
    </Dialog>
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
