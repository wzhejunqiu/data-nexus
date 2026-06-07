import { create } from 'zustand'

export interface ToastItem {
  id: string
  message: string
  variant?: 'error' | 'info'
}

interface ToastState {
  items: ToastItem[]
  push: (message: string, variant?: ToastItem['variant']) => void
  dismiss: (id: string) => void
}

export const useToastStore = create<ToastState>((set, get) => ({
  items: [],
  push: (message, variant = 'info') => {
    const id = `${Date.now()}-${Math.random()}`
    set({ items: [...get().items, { id, message, variant }] })
    setTimeout(() => get().dismiss(id), 4000)
  },
  dismiss: (id) => set({ items: get().items.filter((t) => t.id !== id) }),
}))

export function Toaster() {
  const items = useToastStore((s) => s.items)
  const dismiss = useToastStore((s) => s.dismiss)
  if (items.length === 0) return null
  return (
    <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
      {items.map((item) => (
        <div
          key={item.id}
          className={`rounded border px-3 py-2 text-sm shadow ${
            item.variant === 'error'
              ? 'border-red-500/40 bg-red-500/10 text-red-500'
              : 'border-border bg-card text-foreground'
          }`}
          onClick={() => dismiss(item.id)}
        >
          {item.message}
        </div>
      ))}
    </div>
  )
}
