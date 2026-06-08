import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import type { DriverType } from '@/lib/types'

const DIALECTS: DriverType[] = ['sqlite', 'postgres', 'mysql']

export function DialectSidebar({
  dialect,
  onDialectChange,
  disabled,
}: {
  dialect: DriverType
  onDialectChange: (d: DriverType) => void
  disabled?: boolean
}) {
  const { t } = useTranslation()
  return (
    <nav className="flex w-36 shrink-0 flex-col gap-1 border-r border-border pr-3">
      {DIALECTS.map((d) => (
        <button
          key={d}
          type="button"
          disabled={disabled}
          onClick={() => onDialectChange(d)}
          className={cn(
            'rounded px-2 py-1.5 text-left text-sm transition-colors',
            dialect === d
              ? 'bg-accent/15 font-medium text-foreground'
              : 'text-muted hover:bg-muted/10 hover:text-foreground',
            disabled && 'cursor-not-allowed opacity-50',
          )}
        >
          {t(`connection.type.${d}`)}
        </button>
      ))}
    </nav>
  )
}
