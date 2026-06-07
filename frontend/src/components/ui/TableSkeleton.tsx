export function TableSkeleton({ columns = 4, rows = 8 }: { columns?: number; rows?: number }) {
  return (
    <div className="p-4" data-testid="table-skeleton">
      <div className="overflow-hidden rounded border border-border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border">
              {Array.from({ length: columns }, (_, i) => (
                <th key={i} className="px-3 py-2">
                  <div className="h-4 w-20 animate-pulse rounded bg-muted/40" />
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {Array.from({ length: rows }, (_, row) => (
              <tr key={row} className="border-b border-border/40">
                {Array.from({ length: columns }, (_, col) => (
                  <td key={col} className="px-3 py-2">
                    <div
                      className="h-3 animate-pulse rounded bg-muted/40"
                      style={{ width: `${50 + ((row + col) % 4) * 15}%` }}
                    />
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
