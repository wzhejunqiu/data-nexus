import { cn } from '@/lib/utils'
import type { InputHTMLAttributes } from 'react'

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        'w-full rounded border border-border bg-transparent px-3 py-2 text-sm',
        'focus:outline-none focus:ring-2 focus:ring-accent/40',
        className,
      )}
      {...props}
    />
  )
}
