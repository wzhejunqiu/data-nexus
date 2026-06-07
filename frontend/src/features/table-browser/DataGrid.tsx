import { flexRender, getCoreRowModel, useReactTable, type ColumnDef } from '@tanstack/react-table'
import { useVirtualizer } from '@tanstack/react-virtual'
import { useQuery } from '@tanstack/react-query'
import { useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { formatError } from '@/lib/api/errors'
import { tableApi } from '@/lib/api/table'
import { formatCell } from '@/lib/utils'
import { Pagination } from './Pagination'
import { useStatusStore } from '@/stores/statusStore'

const PAGE_SIZES = [25, 50, 100, 200]

export function DataGrid({ connectionId, tableName }: { connectionId: string; tableName: string }) {
  const { t } = useTranslation()
  const setStatus = useStatusStore((s) => s.setStatus)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [sort, setSort] = useState('')
  const [order, setOrder] = useState<'asc' | 'desc'>('asc')
  const parentRef = useRef<HTMLDivElement>(null)

  const { data, isLoading, error, isFetching } = useQuery({
    queryKey: ['rows', connectionId, tableName, page, pageSize, sort, order],
    queryFn: async () => {
      const result = await tableApi.browseRows({
        connectionId,
        tableName,
        page,
        pageSize,
        sort,
        order,
      })
      setStatus('browse', result.pagination.totalRows, null)
      return result
    },
  })

  const columns = useMemo<ColumnDef<Record<string, unknown>>[]>(() => {
    if (!data) return []
    return data.columns.map((col) => ({
      id: col.name,
      accessorKey: col.name,
      header: col.name,
      cell: ({ getValue }) => {
        const val = getValue()
        const isNull = val === null || val === undefined
        return (
          <span className={isNull ? 'italic text-muted' : ''}>
            {isNull ? t('data.null') : formatCell(val)}
          </span>
        )
      },
    }))
  }, [data, t])

  const table = useReactTable({
    data: data?.rows ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  const rows = table.getRowModel().rows
  const useVirtual = rows.length > 200
  const virtualizer = useVirtualizer({
    count: rows.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 32,
    overscan: 10,
  })

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

  if (isLoading) return <p className="p-4 text-sm text-muted">{t('common.loading')}</p>
  if (error) return <p className="p-4 text-sm text-red-500">{formatError(t, error)}</p>
  if (!data) return null

  const columnNames = data.columns.map((c) => c.name)

  return (
    <div className="flex h-full flex-col p-4">
      <div className="mb-3 flex items-center gap-3 text-sm">
        <label className="flex items-center gap-2">
          {t('data.pageSize')}
          <select
            className="rounded border border-border bg-transparent px-2 py-1"
            value={pageSize}
            onChange={(e) => {
              setPageSize(Number(e.target.value))
              setPage(1)
            }}
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
      <div ref={parentRef} className="min-h-0 flex-1 overflow-auto rounded border border-border">
        <table className="w-full text-sm">
          <thead className="sticky top-0 bg-card">
            <tr>
              {columnNames.map((col) => (
                <th
                  key={col}
                  className="cursor-pointer border-b border-border px-3 py-2 text-left font-medium hover:bg-muted/30"
                  onClick={() => toggleSort(col)}
                >
                  {col}
                  {sort === col ? (order === 'asc' ? ' ▲' : ' ▼') : ''}
                </th>
              ))}
            </tr>
          </thead>
          <tbody
            style={
              useVirtual
                ? { height: `${virtualizer.getTotalSize()}px`, position: 'relative' }
                : undefined
            }
          >
            {useVirtual
              ? virtualizer.getVirtualItems().map((virtualRow) => {
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
                })
              : rows.map((row) => (
                  <tr key={row.id} className="border-b border-border/40">
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id} className="px-3 py-2 font-mono text-xs">
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </td>
                    ))}
                  </tr>
                ))}
          </tbody>
        </table>
      </div>
      <Pagination pagination={data.pagination} page={page} onPageChange={setPage} />
    </div>
  )
}
