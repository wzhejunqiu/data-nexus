import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
  type Row,
} from '@tanstack/react-table'
import { useVirtualizer } from '@tanstack/react-virtual'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useMemo, useRef, useState, type RefObject } from 'react'
import { useTranslation } from 'react-i18next'
import { NullCell } from '@/components/ui/NullCell'
import { TableSkeleton } from '@/components/ui/TableSkeleton'
import { Button } from '@/components/ui/Button'
import { useToastStore } from '@/components/ui/Toast'
import { ExportCsvDialog } from '@/features/csv/ExportCsvDialog'
import { ImportWizard } from '@/features/import/ImportWizard'
import { BatchEditConfirmDialog } from './BatchEditConfirmDialog'
import { Pagination } from './Pagination'
import { connectionApi } from '@/lib/api/connection'
import { dialogApi } from '@/lib/api/dialog'
import { exportApi } from '@/lib/api/export'
import { formatError, isDialogCancelled, isExportCancelled, mapWailsError } from '@/lib/api/errors'
import { fileApi } from '@/lib/api/file'
import { schemaApi } from '@/lib/api/schema'
import { tableApi } from '@/lib/api/table'
import type { PaginatedTableData, PendingEdit } from '@/lib/types'
import type { CSVFormatOptions } from '@/lib/types/csv'
import { saveCSVFormatPreference } from '@/lib/types/csv'
import {
  cellKey,
  formatCell,
  isBlobValue,
  rowsToCSV,
} from '@/lib/utils'
import { useStatusStore } from '@/stores/statusStore'

const PAGE_SIZES = [25, 50, 100, 200]
const VIRTUAL_ROW_THRESHOLD = 200
const MAX_BATCH = 200

export function DataGrid({ connectionId, tableName }: { connectionId: string; tableName: string }) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const pushToast = useToastStore((s) => s.push)
  const setStatus = useStatusStore((s) => s.setStatus)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [sort, setSort] = useState('')
  const [order, setOrder] = useState<'asc' | 'desc'>('asc')
  const [pendingEdits, setPendingEdits] = useState<Map<string, PendingEdit>>(new Map())
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [exportOpen, setExportOpen] = useState(false)
  const [exportMode, setExportMode] = useState<'page' | 'all'>('page')
  const [importOpen, setImportOpen] = useState(false)

  const { data: connections } = useQuery({
    queryKey: ['connections'],
    queryFn: () => connectionApi.list(),
  })
  const connItem = connections?.items.find((c) => c.id === connectionId)
  const readOnly = connItem?.config.sqlite?.readOnly ?? false

  const { data: schema } = useQuery({
    queryKey: ['schema', connectionId, tableName],
    queryFn: () => schemaApi.getTableSchema(connectionId, tableName),
  })

  const pkColumns = useMemo(
    () => schema?.columns.filter((c) => c.primaryKey).map((c) => c.name) ?? [],
    [schema],
  )
  const isView = schema?.type === 'view'
  const canEdit = !readOnly && !isView && pkColumns.length > 0

  const { data, isLoading, error, isFetching } = useQuery({
    queryKey: ['rows', connectionId, tableName, page, pageSize, sort, order],
    queryFn: () =>
      tableApi.browseRows({
        connectionId,
        tableName,
        page,
        pageSize,
        sort,
        order,
      }),
  })

  useEffect(() => {
    if (data) {
      setStatus('browse', data.pagination.totalRows, null)
    }
  }, [data, setStatus])

  const saveBatch = useMutation({
    mutationFn: () => {
      const changes = Array.from(pendingEdits.values()).map((e) => ({
        columnName: e.columnName,
        primaryKey: e.primaryKey,
        newValue: e.newValue,
      }))
      return tableApi.updateCellsBatch({ connectionId, tableName, changes })
    },
    onSuccess: (res) => {
      setPendingEdits(new Map())
      setConfirmOpen(false)
      pushToast(t('edit.saveSuccess', { count: res.updatedCount }), 'info')
      void qc.invalidateQueries({ queryKey: ['rows', connectionId, tableName] })
    },
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

  const guardNavigation = useCallback(
    (action: () => void) => {
      if (pendingEdits.size === 0) {
        action()
        return
      }
      if (window.confirm(t('edit.unsavedWarning'))) {
        setPendingEdits(new Map())
        action()
      }
    },
    [pendingEdits.size, t],
  )

  const toggleSort = (col: string) => {
    guardNavigation(() => {
      if (sort !== col) {
        setSort(col)
        setOrder('asc')
      } else if (order === 'asc') {
        setOrder('desc')
      } else {
        setSort('')
        setOrder('asc')
      }
      setPage(1)
    })
  }

  const commitEdit = useCallback(
    (edit: PendingEdit) => {
      const key = cellKey(edit.primaryKey, edit.columnName)
      const same =
        edit.newValue === edit.originalValue ||
        (edit.newValue === '' && (edit.originalValue === null || edit.originalValue === undefined))
      setPendingEdits((prev) => {
        const next = new Map(prev)
        if (same) next.delete(key)
        else next.set(key, edit)
        return next
      })
    },
    [],
  )

  const handleExport = async (format: CSVFormatOptions, exportId: string | null) => {
    saveCSVFormatPreference(format)
    try {
      if (exportMode === 'all') {
        const path = await dialogApi.saveFile(`${tableName}.csv`, [
          { displayName: 'CSV', pattern: '*.csv' },
        ])
        await exportApi.exportTableCSV({
          connectionId,
          tableName,
          format,
          defaultPath: path,
          exportId: exportId ?? undefined,
        })
        pushToast(t('csv.exportSuccess'), 'info')
        return
      }
      if (!data) return
      const columns = data.columns.map((c) => c.name)
      const csv = rowsToCSV(columns, data.rows, format)
      const path = await dialogApi.saveFile(`${tableName}.csv`, [
        { displayName: 'CSV', pattern: '*.csv' },
      ])
      await fileApi.writeTextFile(path, csv, format.encoding)
      pushToast(t('csv.exportSuccess'), 'info')
    } catch (err) {
      const appErr = mapWailsError(err)
      if (isExportCancelled(appErr)) {
        pushToast(t('csv.exportCancelled'), 'info')
      } else if (!isDialogCancelled(appErr)) {
        pushToast(formatError(t, appErr), 'error')
      }
      throw err
    }
  }

  const handleCancelExport = async (exportId: string) => {
    await exportApi.cancelExportTableCSV(exportId)
  }

  if (isLoading) return <TableSkeleton />
  if (error) return <p className="p-4 text-sm text-red-500">{formatError(t, error)}</p>
  if (!data) return null

  return (
    <>
      <DataGridTable
        data={data}
        page={page}
        pageSize={pageSize}
        sort={sort}
        order={order}
        isFetching={isFetching}
        canEdit={canEdit}
        readOnly={readOnly}
        isView={isView}
        hasPK={pkColumns.length > 0}
        pkColumns={pkColumns}
        pendingEdits={pendingEdits}
        submitting={saveBatch.isPending}
        onPageChange={(p) => guardNavigation(() => setPage(p))}
        onPageSizeChange={(size) => guardNavigation(() => { setPageSize(size); setPage(1) })}
        onToggleSort={toggleSort}
        onCommitEdit={commitEdit}
        onDiscard={() => setPendingEdits(new Map())}
        onSave={() => {
          if (pendingEdits.size > MAX_BATCH) {
            pushToast(t('edit.batchLimit', { max: MAX_BATCH }), 'error')
            return
          }
          setConfirmOpen(true)
        }}
        pendingCount={pendingEdits.size}
        onExportPage={() => { setExportMode('page'); setExportOpen(true) }}
        onExportAll={() => { setExportMode('all'); setExportOpen(true) }}
        onImport={() => setImportOpen(true)}
      />
      <BatchEditConfirmDialog
        open={confirmOpen}
        edits={Array.from(pendingEdits.values())}
        onOpenChange={setConfirmOpen}
        onConfirm={() => saveBatch.mutate()}
        busy={saveBatch.isPending}
      />
      <ExportCsvDialog
        open={exportOpen}
        title={exportMode === 'all' ? t('csv.exportAll') : t('csv.exportPage')}
        mode={exportMode}
        totalRows={data.pagination.totalRows}
        onOpenChange={setExportOpen}
        onExport={handleExport}
        onCancelExport={handleCancelExport}
      />
      {importOpen && (
        <ImportWizard
          connectionId={connectionId}
          defaultTable={tableName}
          onClose={() => setImportOpen(false)}
          onDone={() => {
            setImportOpen(false)
            void qc.invalidateQueries({ queryKey: ['rows', connectionId, tableName] })
            void qc.invalidateQueries({ queryKey: ['schema', connectionId] })
          }}
        />
      )}
    </>
  )
}

function DataGridTable({
  data,
  page,
  pageSize,
  sort,
  order,
  isFetching,
  canEdit,
  readOnly,
  isView,
  hasPK,
  pkColumns,
  pendingEdits,
  submitting,
  pendingCount,
  onPageChange,
  onPageSizeChange,
  onToggleSort,
  onCommitEdit,
  onDiscard,
  onSave,
  onExportPage,
  onExportAll,
  onImport,
}: {
  data: PaginatedTableData
  page: number
  pageSize: number
  sort: string
  order: 'asc' | 'desc'
  isFetching: boolean
  canEdit: boolean
  readOnly: boolean
  isView: boolean
  hasPK: boolean
  pkColumns: string[]
  pendingEdits: Map<string, PendingEdit>
  submitting: boolean
  pendingCount: number
  onPageChange: (page: number) => void
  onPageSizeChange: (size: number) => void
  onToggleSort: (col: string) => void
  onCommitEdit: (edit: PendingEdit) => void
  onDiscard: () => void
  onSave: () => void
  onExportPage: () => void
  onExportAll: () => void
  onImport: () => void
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const parentRef = useRef<HTMLDivElement>(null)
  const [editing, setEditing] = useState<{ rowIdx: number; col: string } | null>(null)
  const [editValue, setEditValue] = useState('')

  const columnNames = data.columns.map((c) => c.name)
  const blobCols = useMemo(
    () => new Set(data.columns.filter((c) => c.dataType.toUpperCase().includes('BLOB')).map((c) => c.name)),
    [data.columns],
  )

  const getPK = useCallback(
    (row: Record<string, unknown>) => {
      const pk: Record<string, unknown> = {}
      for (const k of pkColumns) pk[k] = row[k]
      return pk
    },
    [pkColumns],
  )

  const startEdit = (rowIdx: number, col: string, row: Record<string, unknown>) => {
    if (readOnly) {
      pushToast(t('sql.readOnlyBlocked'), 'error')
      return
    }
    if (isView) {
      pushToast(t('edit.noView'), 'error')
      return
    }
    if (!hasPK) {
      pushToast(t('edit.noPK'), 'error')
      return
    }
    if (blobCols.has(col)) {
      pushToast(t('edit.noBlob'), 'error')
      return
    }
    const pk = getPK(row)
    const key = cellKey(pk, col)
    const pending = pendingEdits.get(key)
    const val = pending ? pending.newValue : row[col]
    setEditing({ rowIdx, col })
    setEditValue(val === null || val === undefined ? '' : String(val))
  }

  const finishEdit = (rowIdx: number, col: string, move: 'none' | 'down' | 'next' | 'prev') => {
    const row = data.rows[rowIdx]
    if (!row) return
    const pk = getPK(row)
    const original = row[col]
    const newValue = editValue === '' ? null : editValue
    onCommitEdit({ columnName: col, primaryKey: pk, originalValue: original, newValue })
    setEditing(null)
    if (move === 'down') {
      const nextRow = rowIdx + 1
      if (nextRow < data.rows.length) startEdit(nextRow, col, data.rows[nextRow])
    } else if (move === 'next' || move === 'prev') {
      const cols = columnNames.filter((c) => !blobCols.has(c))
      const idx = cols.indexOf(col)
      const nextIdx = move === 'next' ? idx + 1 : idx - 1
      if (nextIdx >= 0 && nextIdx < cols.length) {
        startEdit(rowIdx, cols[nextIdx], row)
      }
    }
  }

  const columns = useMemo<ColumnDef<Record<string, unknown>>[]>(
    () =>
      data.columns.map((col) => ({
        id: col.name,
        accessorKey: col.name,
        header: col.name,
        cell: ({ row, getValue }) => {
          const rowIdx = row.index
          const val = getValue()
          const pk = getPK(row.original)
          const key = cellKey(pk, col.name)
          const pending = pendingEdits.get(key)
          const displayVal = pending ? pending.newValue : val
          const isDirty = !!pending

          if (editing?.rowIdx === rowIdx && editing.col === col.name) {
            return (
              <input
                autoFocus
                className="w-full min-w-[80px] rounded border border-accent bg-background px-1 py-0.5 font-mono text-xs"
                value={editValue}
                onChange={(e) => setEditValue(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    finishEdit(rowIdx, col.name, 'down')
                  } else if (e.key === 'Tab') {
                    e.preventDefault()
                    finishEdit(rowIdx, col.name, e.shiftKey ? 'prev' : 'next')
                  } else if (e.key === 'Escape') {
                    setEditing(null)
                  }
                }}
                onBlur={() => finishEdit(rowIdx, col.name, 'none')}
                disabled={submitting}
              />
            )
          }

          const isNull = displayVal === null || displayVal === undefined
          return (
            <span
              className={isDirty ? 'border-l-2 border-amber-500 pl-1' : undefined}
              onDoubleClick={() => canEdit && startEdit(rowIdx, col.name, row.original)}
              title={canEdit ? t('edit.doubleClick') : undefined}
            >
              {isNull ? (
                <NullCell label={t('data.null')} />
              ) : isBlobValue(displayVal) ? (
                formatCell(displayVal)
              ) : (
                formatCell(displayVal)
              )}
            </span>
          )
        },
      })),
    [data.columns, pendingEdits, editing, editValue, canEdit, getPK, t, submitting],
  )

  const table = useReactTable({
    data: data.rows,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  const rows = table.getRowModel().rows
  const isEmpty = data.rows.length === 0
  const useVirtual = rows.length > VIRTUAL_ROW_THRESHOLD

  return (
    <div className="flex h-full flex-col p-4">
      <div className="mb-3 flex flex-wrap items-center gap-2 text-sm">
        <label className="flex items-center gap-2">
          {t('data.pageSize')}
          <select
            className="rounded border border-border bg-transparent px-2 py-1"
            value={pageSize}
            onChange={(e) => onPageSizeChange(Number(e.target.value))}
          >
            {PAGE_SIZES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </label>
        <span className="text-muted">{t('data.total', { count: data.pagination.totalRows })}</span>
        {isFetching && <span className="text-muted">{t('common.loading')}</span>}
        <div className="ml-auto flex flex-wrap gap-2">
          <Button size="sm" variant="outline" onClick={onExportPage}>
            {t('csv.exportPage')}
          </Button>
          <Button size="sm" variant="outline" onClick={onExportAll}>
            {t('csv.exportAll')}
          </Button>
          {!readOnly && (
            <Button size="sm" variant="outline" onClick={onImport}>
              {t('csv.import')}
            </Button>
          )}
          {pendingCount > 0 && (
            <>
              <Button size="sm" onClick={onSave} disabled={submitting}>
                {t('edit.saveChanges', { count: pendingCount })}
              </Button>
              <Button size="sm" variant="outline" onClick={onDiscard} disabled={submitting}>
                {t('edit.discardChanges')}
              </Button>
            </>
          )}
        </div>
      </div>
      {isEmpty ? (
        <p className="mt-3 text-sm text-muted">{t('data.emptyTable')}</p>
      ) : (
        <div ref={parentRef} className="min-h-0 flex-1 overflow-auto rounded border border-border">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-card">
              <tr>
                {columnNames.map((col) => (
                  <th
                    key={col}
                    className="cursor-pointer border-b border-border px-3 py-2 text-left font-medium hover:bg-muted/30"
                    onClick={() => onToggleSort(col)}
                  >
                    {col}
                    {sort === col ? (order === 'asc' ? ' ▲' : ' ▼') : ''}
                  </th>
                ))}
              </tr>
            </thead>
            {useVirtual ? (
              <VirtualizedTableBody parentRef={parentRef} rows={rows} />
            ) : (
              <tbody>
                {rows.map((row) => (
                  <tr key={row.id} className="border-b border-border/40">
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id} className="px-3 py-2 font-mono text-xs">
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            )}
          </table>
        </div>
      )}
      <Pagination pagination={data.pagination} page={page} onPageChange={onPageChange} />
    </div>
  )
}

function VirtualizedTableBody({
  parentRef,
  rows,
}: {
  parentRef: RefObject<HTMLDivElement | null>
  rows: Row<Record<string, unknown>>[]
}) {
  const virtualizer = useVirtualizer({
    count: rows.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 32,
    overscan: 10,
  })

  return (
    <tbody style={{ height: `${virtualizer.getTotalSize()}px`, position: 'relative' }}>
      {virtualizer.getVirtualItems().map((virtualRow) => {
        const row = rows[virtualRow.index]
        return (
          <tr
            key={row.id}
            className="border-b border-border/40"
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              transform: `translateY(${virtualRow.start}px)`,
            }}
          >
            {row.getVisibleCells().map((cell) => (
              <td key={cell.id} className="px-3 py-2 font-mono text-xs">
                {flexRender(cell.column.columnDef.cell, cell.getContext())}
              </td>
            ))}
          </tr>
        )
      })}
    </tbody>
  )
}
