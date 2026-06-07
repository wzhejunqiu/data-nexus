export function NullCell({ label }: { label: string }) {
  return (
    <span className="inline-block rounded bg-muted/50 px-1.5 py-0.5 font-mono text-xs italic text-muted">
      {label}
    </span>
  )
}
