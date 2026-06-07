export function ProgressBar({
  value,
  max,
  indeterminate,
  className,
}: {
  value?: number
  max?: number
  indeterminate?: boolean
  className?: string
}) {
  if (indeterminate) {
    return (
      <div className={className}>
        <progress className="h-2 w-full overflow-hidden rounded-full accent-blue-600" />
      </div>
    )
  }
  const safeValue = value ?? 0
  const safeMax = max != null && max > 0 ? max : 1
  const pct = Math.min(100, Math.round((safeValue / safeMax) * 100))
  return (
    <div className={className}>
      <progress
        className="h-2 w-full overflow-hidden rounded-full accent-blue-600"
        value={safeValue}
        max={safeMax}
      />
      <p className="mt-1 text-xs text-muted">{pct}%</p>
    </div>
  )
}
