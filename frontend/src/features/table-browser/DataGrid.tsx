import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
  type Row,
} from '@tanstack/react-table'
import { useVirtualizer } from '@tanstack/react-virtual'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useMemo, useRef, useState, type RefObject } from 'react'
import { useTranslation } from 'react-i18next'
import { NullCell } from '@/components/ui/NullCell'
import { TableSkeleton } from '@/components/ui/TableSkeleton'
import { formatError } from '@/lib/api/errors'
import { tableApi } from '@/lib/api/table'
import type { PaginatedTableData } from '@/lib/types'
import { formatCell } from '@/lib/utils'
import { Pagination } from './Pagination'
import { useStatusStore } from '@/stores/statusStore'

const PAGE_SIZES = [25, 50, 100, 200]
const VIRTUAL_ROW_THRESHOLD = 200

export function DataGrid({ connectionId, tableName }: { connectionId: string; tableName: string }) {
  const { t } = useTranslation()
  const setStatus = useStatusStore((s) => s.setStatus)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [sort, setSort] = useState('')
  const [order, setOrder] = useState<'asc' | 'desc'>('asc')

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

  const toggleSort = (col: string) => {
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
  }

  if (isLoading) return <TableSkeleton />
  if (error) return <p className="p-4 text-sm text-red-500">{formatError(t, error)}</p>
  if (!data) return null

  return (
    <DataGridTable
      data={data}
      page={page}
      pageSize={pageSize}
      sort={sort}
      order={order}
      isFetching={isFetching}
      onPageChange={setPage}
      onPageSizeChange={(size) => {
        setPageSize(size)
        setPage(1)
      }}
      onToggleSort={toggleSort}
    />
  )
}

function DataGridTable({
  data,
  page,
  pageSize,
  sort,
  order,
  isFetching,
  onPageChange,
  onPageSizeChange,
  onToggleSort,
}: {
  data: PaginatedTableData
  page: number
  pageSize: number
  sort: string
  order: 'asc' | 'desc'
  isFetching: boolean
  onPageChange: (page: number) => void
  onPageSizeChange: (size: number) => void
  onToggleSort: (col: string) => void
}) {
  const { t } = useTranslation()
  const parentRef = useRef<HTMLDivElement>(null)

  const columns = useMemo<ColumnDef<Record<string, unknown>>[]>(
    () =>
      data.columns.map((col) => ({
        id: col.name,
        accessorKey: col.name,
        header: col.name,
        cell: ({ getValue }) => {
          const val = getValue()
          const isNull = val === null || val === undefined
          return isNull ? <NullCell label={t('data.null')} /> : formatCell(val)
        },
      })),
    [data.columns, t],
  )

  const table = useReactTable({
    data: data.rows,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  const rows = table.getRowModel().rows
  const columnNames = data.columns.map((c) => c.name)
  const isEmpty = data.rows.length === 0
  const useVirtual = rows.length > VIRTUAL_ROW_THRESHOLD

  return (
    <div className="flex h-full flex-col p-4">
      <div className="mb-3 flex items-center gap-3 text-sm">
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
