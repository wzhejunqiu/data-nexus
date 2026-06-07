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
import { useExportStore } from '@/stores/exportStore'
import { useImportStore } from '@/stores/importStore'
import { BatchEditConfirmDialog } from './BatchEditConfirmDialog'
import { Pagination } from './Pagination'
import { FilterBuilder } from './FilterBuilder'
import { FtsSearchBar } from './FtsSearchBar'
import { FacetPanel } from './FacetPanel'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import type { RowFilter } from '@/lib/types'
import { connectionApi } from '@/lib/api/connection'
import { formatError } from '@/lib/api/errors'
import { schemaApi } from '@/lib/api/schema'
import { tableApi } from '@/lib/api/table'
import type { PaginatedTableData, PendingEdit } from '@/lib/types'
import { cellKey, coerceNewValue, formatCell, isBlobValue, valuesEqual } from '@/lib/utils'
import { isBatchOverLimit, MAX_BATCH_EDITS } from '@/lib/editBatch'
import { useStatusStore } from '@/stores/statusStore'

const PAGE_SIZES = [25, 50, 100, 200]
const VIRTUAL_ROW_THRESHOLD = 200

export function DataGrid({ connectionId, tableName }: { connectionId: string; tableName: string }) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const pushToast = useToastStore((s) => s.push)
  const setStatus = useStatusStore((s) => s.setStatus)
  const getTableState = useWorkspaceStore((s) => s.getTableState)
  const setTableState = useWorkspaceStore((s) => s.setTableState)
  const saved = getTableState(connectionId, tableName)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [sort, setSort] = useState(saved.sort)
  const [order, setOrder] = useState<'asc' | 'desc'>(saved.order)
  const [filters, setFilters] = useState<RowFilter[]>(saved.filters)
  const [search, setSearch] = useState(saved.search)
  const [appliedFilters, setAppliedFilters] = useState<RowFilter[]>(saved.filters)
  const [appliedSearch, setAppliedSearch] = useState(saved.search)
  const [pendingEdits, setPendingEdits] = useState<Map<string, PendingEdit>>(new Map())
  const [confirmOpen, setConfirmOpen] = useState(false)
  const openExport = useExportStore((s) => s.openExport)
  const openImport = useImportStore((s) => s.openImport)

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

  const { data: ftsInfo } = useQuery({
    queryKey: ['fts', connectionId, tableName],
    queryFn: () => schemaApi.detectFTS(connectionId, tableName),
  })

  useEffect(() => {
    setTableState(connectionId, tableName, {
      sort,
      order,
      filters: appliedFilters,
      search: appliedSearch,
    })
  }, [connectionId, tableName, sort, order, appliedFilters, appliedSearch, setTableState])

  const pkColumns = useMemo(
    () => schema?.columns.filter((c) => c.primaryKey).map((c) => c.name) ?? [],
    [schema],
  )
  const isView = schema?.type === 'view'

  const { data, isLoading, error, isFetching } = useQuery({
    queryKey: [
      'rows',
      connectionId,
      tableName,
      page,
      pageSize,
      sort,
      order,
      appliedFilters,
      appliedSearch,
    ],
    queryFn: () =>
      tableApi.browseRows({
        connectionId,
        tableName,
        page,
        pageSize,
        sort,
        order,
        filters: appliedFilters,
        search: appliedSearch || undefined,
      }),
  })

  const usesRowidKey =
    pkColumns.length === 0 && !isView && (data?.rows.some((r) => r.rowid != null) ?? false)
  const canEdit = !readOnly && !isView && (pkColumns.length > 0 || usesRowidKey)

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

  const commitEdit = useCallback((edit: PendingEdit) => {
    const key = cellKey(edit.primaryKey, edit.columnName)
    const same = valuesEqual(edit.originalValue, edit.newValue)
    setPendingEdits((prev) => {
      const next = new Map(prev)
      if (same) next.delete(key)
      else next.set(key, edit)
      return next
    })
  }, [])

  const openTableExport = () => {
    if (!data) return
    openExport({
      source: 'table-page',
      connectionId,
      tableName,
      availableColumns: data.columns.map((c) => c.name),
      rows: data.rows,
      totalRows: data.pagination.totalRows,
    })
  }

  if (isLoading) return <TableSkeleton />
  if (error) return <p className="p-4 text-sm text-red-500">{formatError(t, error)}</p>
  if (!data) return null

  const applyFilters = () => {
    guardNavigation(() => {
      setAppliedFilters(filters.filter((f) => f.column))
      setAppliedSearch(search)
      setPage(1)
    })
  }

  const toggleFacetFilter = (filter: RowFilter) => {
    const isActive = appliedFilters.some(
      (f) => f.column === filter.column && f.operator === 'eq' && f.value === filter.value,
    )
    const next = isActive
      ? filters.filter(
          (f) => !(f.column === filter.column && f.operator === 'eq' && f.value === filter.value),
        )
      : [...filters.filter((f) => !(f.column === filter.column && f.operator === 'eq')), filter]
    setFilters(next)
    setAppliedFilters(next)
    setPage(1)
  }

  return (
    <>
      {schema && (
        <>
          <FacetPanel
            connectionId={connectionId}
            tableName={tableName}
            filters={appliedFilters}
            onToggleFilter={toggleFacetFilter}
          />
          <FilterBuilder
            tableName={tableName}
            columns={schema.columns}
            filters={filters}
            sort={sort}
            order={order}
            onChange={setFilters}
            onApply={applyFilters}
          />
        </>
      )}
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
        usesRowidKey={usesRowidKey}
        pkColumns={pkColumns}
        pendingEdits={pendingEdits}
        submitting={saveBatch.isPending}
        onPageChange={(p) => guardNavigation(() => setPage(p))}
        onPageSizeChange={(size) =>
          guardNavigation(() => {
            setPageSize(size)
            setPage(1)
          })
        }
        onToggleSort={toggleSort}
        onCommitEdit={commitEdit}
        onDiscard={() => setPendingEdits(new Map())}
        onSave={() => {
          if (isBatchOverLimit(pendingEdits.size)) {
            pushToast(t('edit.batchLimit', { max: MAX_BATCH_EDITS }), 'error')
            return
          }
          setConfirmOpen(true)
        }}
        pendingCount={pendingEdits.size}
        onExportCsv={openTableExport}
        onImport={() => openImport({ connectionId, defaultTable: tableName })}
        ftsEnabled={ftsInfo?.enabled ?? false}
        ftsBarKey={`${connectionId}:${tableName}`}
        search={search}
        onSearchChange={setSearch}
        onSearchApply={() =>
          guardNavigation(() => {
            setAppliedSearch(search)
            setPage(1)
          })
        }
      />
      <BatchEditConfirmDialog
        open={confirmOpen}
        edits={Array.from(pendingEdits.values())}
        onOpenChange={setConfirmOpen}
        onConfirm={() => saveBatch.mutate()}
        busy={saveBatch.isPending}
      />
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
  usesRowidKey,
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
  onExportCsv,
  onImport,
  ftsEnabled,
  ftsBarKey,
  search,
  onSearchChange,
  onSearchApply,
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
  usesRowidKey: boolean
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
  onExportCsv: () => void
  onImport: () => void
  ftsEnabled: boolean
  ftsBarKey: string
  search: string
  onSearchChange: (v: string) => void
  onSearchApply: () => void
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const parentRef = useRef<HTMLDivElement>(null)
  const cancelEditRef = useRef(false)
  const [editing, setEditing] = useState<{ rowIdx: number; col: string } | null>(null)
  const [editValue, setEditValue] = useState('')

  const columnNames = data.columns.map((c) => c.name)
  const blobCols = useMemo(
    () =>
      new Set(
        data.columns.filter((c) => c.dataType.toUpperCase().includes('BLOB')).map((c) => c.name),
      ),
    [data.columns],
  )
  const systemCols = useMemo(() => {
    const cols = new Set(blobCols)
    if (usesRowidKey) cols.add('rowid')
    return cols
  }, [blobCols, usesRowidKey])

  const getPK = useCallback(
    (row: Record<string, unknown>) => {
      if (pkColumns.length > 0) {
        const pk: Record<string, unknown> = {}
        for (const k of pkColumns) pk[k] = row[k]
        return pk
      }
      return { rowid: row.rowid }
    },
    [pkColumns],
  )

  const startEdit = useCallback(
    (rowIdx: number, col: string, row: Record<string, unknown>) => {
      if (readOnly) {
        pushToast(t('sql.readOnlyBlocked'), 'error')
        return
      }
      if (isView) {
        pushToast(t('edit.noView'), 'error')
        return
      }
      if (!hasPK && !usesRowidKey) {
        pushToast(t('edit.noPK'), 'error')
        return
      }
      if (systemCols.has(col)) {
        if (col === 'rowid') pushToast(t('edit.noRowid'), 'error')
        else pushToast(t('edit.noBlob'), 'error')
        return
      }
      const pk = getPK(row)
      const key = cellKey(pk, col)
      const pending = pendingEdits.get(key)
      const val = pending ? pending.newValue : row[col]
      setEditing({ rowIdx, col })
      setEditValue(val === null || val === undefined ? '' : String(val))
    },
    [readOnly, pushToast, t, isView, hasPK, usesRowidKey, systemCols, getPK, pendingEdits],
  )

  const cancelEdit = useCallback(() => {
    cancelEditRef.current = true
    setEditing(null)
  }, [])

  useEffect(() => {
    if (!editing) return
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault()
        cancelEdit()
      }
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [editing, cancelEdit])

  const finishEdit = useCallback(
    (rowIdx: number, col: string, move: 'none' | 'down' | 'next' | 'prev') => {
      const row = data.rows[rowIdx]
      if (!row) return
      const pk = getPK(row)
      const original = row[col]
      const colMeta = data.columns.find((c) => c.name === col)
      const newValue = coerceNewValue(original, editValue, colMeta?.dataType ?? 'TEXT')
      onCommitEdit({ columnName: col, primaryKey: pk, originalValue: original, newValue })
      setEditing(null)
      if (move === 'down') {
        const nextRow = rowIdx + 1
        if (nextRow < data.rows.length) startEdit(nextRow, col, data.rows[nextRow])
      } else if (move === 'next' || move === 'prev') {
        const cols = columnNames.filter((c) => !systemCols.has(c))
        const idx = cols.indexOf(col)
        const nextIdx = move === 'next' ? idx + 1 : idx - 1
        if (nextIdx >= 0 && nextIdx < cols.length) {
          startEdit(rowIdx, cols[nextIdx], row)
        }
      }
    },
    [columnNames, data.columns, data.rows, editValue, getPK, onCommitEdit, systemCols, startEdit],
  )

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
                    e.preventDefault()
                    cancelEdit()
                  }
                }}
                onBlur={() => {
                  if (cancelEditRef.current) {
                    cancelEditRef.current = false
                    return
                  }
                  finishEdit(rowIdx, col.name, 'none')
                }}
                disabled={submitting}
              />
            )
          }

          const isNull = displayVal === null || displayVal === undefined
          const isBlobCol = blobCols.has(col.name)
          return (
            <span
              className={
                [
                  isDirty ? 'border-l-2 border-amber-500 pl-1' : undefined,
                  isBlobCol && canEdit ? 'opacity-60 cursor-not-allowed' : undefined,
                ]
                  .filter(Boolean)
                  .join(' ') || undefined
              }
              onDoubleClick={() => startEdit(rowIdx, col.name, row.original)}
              title={
                isBlobCol && canEdit
                  ? t('edit.noBlob')
                  : canEdit
                    ? t('edit.doubleClick')
                    : undefined
              }
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
    [
      data.columns,
      pendingEdits,
      editing,
      editValue,
      canEdit,
      blobCols,
      getPK,
      t,
      submitting,
      finishEdit,
      cancelEdit,
      startEdit,
    ],
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
        {ftsEnabled && (
          <FtsSearchBar
            key={ftsBarKey}
            defaultValue={search}
            onChange={(v) => {
              onSearchChange(v)
              onSearchApply()
            }}
          />
        )}
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
          <Button size="sm" variant="outline" onClick={onExportCsv}>
            {t('csv.export')}
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
