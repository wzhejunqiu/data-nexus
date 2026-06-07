import { cn } from '@/lib/utils'
import type { ButtonHTMLAttributes } from 'react'

export function Button({
  className,
  variant = 'default',
  size = 'md',
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'default' | 'ghost' | 'danger' | 'outline'
  size?: 'sm' | 'md'
}) {
  return (
    <button
      className={cn(
        'inline-flex items-center justify-center rounded-md font-medium transition-colors disabled:opacity-50',
        size === 'sm' ? 'h-7 px-2 text-xs' : 'h-8 px-3 text-sm',
        variant === 'default' && 'bg-accent text-white hover:opacity-90',
        variant === 'ghost' && 'hover:bg-muted/50',
        variant === 'outline' && 'border border-border hover:bg-muted/30',
        variant === 'danger' && 'text-red-500 hover:bg-red-500/10',
        className,
      )}
      {...props}
    />
  )
}
