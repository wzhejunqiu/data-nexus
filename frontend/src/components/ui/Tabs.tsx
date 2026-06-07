import { cn } from '@/lib/utils'

export function Tabs({
  tabs,
  active,
  onChange,
}: {
  tabs: { id: string; label: string }[]
  active: string
  onChange: (id: string) => void
}) {
  return (
    <div className="flex gap-1 border-b border-border px-2">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          className={cn(
            'px-3 py-2 text-sm border-b-2 -mb-px',
            active === tab.id
              ? 'border-accent text-foreground'
              : 'border-transparent text-muted hover:text-foreground',
          )}
          onClick={() => onChange(tab.id)}
        >
          {tab.label}
        </button>
      ))}
    </div>
  )
}
