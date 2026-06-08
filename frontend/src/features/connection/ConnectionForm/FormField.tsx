import { cn } from '@/lib/utils'
import type { ReactNode } from 'react'

export function FormField({
  id,
  label,
  required,
  optional,
  error,
  children,
}: {
  id: string
  label: string
  required?: boolean
  optional?: string
  error?: string
  children: ReactNode
}) {
  const errorId = error ? `${id}-error` : undefined
  return (
    <div className="space-y-1">
      <label htmlFor={id} className="block text-xs font-medium text-foreground">
        {label}
        {required && (
          <span className="ml-0.5 text-red-500" aria-hidden="true">
            *
          </span>
        )}
        {optional && <span className="ml-1 font-normal text-muted">({optional})</span>}
      </label>
      {children}
      {error && (
        <p id={errorId} className="text-xs text-red-500">
          {error}
        </p>
      )}
    </div>
  )
}

export function fieldErrorClass(hasError: boolean) {
  return cn(hasError && 'border-red-500 focus:ring-red-500/40')
}
