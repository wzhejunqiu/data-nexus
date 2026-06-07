import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { tableApi } from '@/lib/api/table'
import { formatCell } from '@/lib/utils'

const PAGE_SIZES = [25, 50, 100, 200]

export function DataGrid({ connectionId, tableName }: { connectionId: string; tableName: string }) {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [sort, setSort] = useState('')
  const [order, setOrder] = useState<'asc' | 'desc'>('asc')

  const { data, isLoading, error } = useQuery({
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
  if (error) return <p className="p-4 text-sm text-red-500">{(error as Error).message}</p>
  if (!data) return null

  const columns = data.columns.map((c) => c.name)
  const { pagination } = data

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
        <span className="text-muted">{t('data.total', { count: pagination.totalRows })}</span>
      </div>
      <div className="min-h-0 flex-1 overflow-auto rounded border border-border">
        <table className="w-full text-sm">
          <thead className="sticky top-0 bg-card">
            <tr>
              {columns.map((col) => (
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
          <tbody>
            {data.rows.map((row, i) => (
              <tr key={i} className="border-b border-border/40">
                {columns.map((col) => {
                  const val = row[col]
                  const isNull = val === null || val === undefined
                  return (
                    <td
                      key={col}
                      className={`px-3 py-2 font-mono text-xs ${isNull ? 'italic text-muted' : ''}`}
                    >
                      {isNull ? t('data.null') : formatCell(val)}
                    </td>
                  )
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="mt-3 flex items-center gap-2">
        <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage(1)}>
          «
        </Button>
        <Button
          size="sm"
          variant="outline"
          disabled={page <= 1}
          onClick={() => setPage((p) => p - 1)}
        >
          ‹
        </Button>
        <span className="text-sm">
          {pagination.page} / {pagination.totalPages}
        </span>
        <Button
          size="sm"
          variant="outline"
          disabled={page >= pagination.totalPages}
          onClick={() => setPage((p) => p + 1)}
        >
          ›
        </Button>
        <Button
          size="sm"
          variant="outline"
          disabled={page >= pagination.totalPages}
          onClick={() => setPage(pagination.totalPages)}
        >
          »
        </Button>
      </div>
    </div>
  )
}
