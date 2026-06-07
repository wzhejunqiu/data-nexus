import { Button } from '@/components/ui/Button'
import type { PaginationMeta } from '@/lib/types'

export function Pagination({
  pagination,
  page,
  onPageChange,
}: {
  pagination: PaginationMeta
  page: number
  onPageChange: (page: number) => void
}) {
  return (
    <div className="mt-3 flex items-center gap-2">
      <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => onPageChange(1)}>
        «
      </Button>
      <Button
        size="sm"
        variant="outline"
        disabled={page <= 1}
        onClick={() => onPageChange(page - 1)}
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
        onClick={() => onPageChange(page + 1)}
      >
        ›
      </Button>
      <Button
        size="sm"
        variant="outline"
        disabled={page >= pagination.totalPages}
        onClick={() => onPageChange(pagination.totalPages)}
      >
        »
      </Button>
    </div>
  )
}
