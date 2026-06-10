export function TreeChevron({
  open,
  onToggle,
}: {
  open: boolean
  onToggle: (e: React.MouseEvent) => void
}) {
  return (
    <button
      type="button"
      className="shrink-0 px-0.5 text-muted hover:text-foreground"
      aria-label={open ? 'collapse' : 'expand'}
      onClick={onToggle}
    >
      {open ? '▼' : '▶'}
    </button>
  )
}
