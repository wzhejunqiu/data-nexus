import { AppMenuBar, type AppMenuBarProps, useAppShortcuts } from '@/components/AppMenuBar'
import { useIsMacOS } from '@/lib/platform'

export function Header(props: AppMenuBarProps) {
  const isMacOS = useIsMacOS()
  return (
    <header className={`flex items-center border-b border-border px-4 ${isMacOS ? 'h-9' : 'h-12'}`}>
      <AppMenuBar {...props} />
    </header>
  )
}

export function StatusBar({ text }: { text: string }) {
  return (
    <footer className="flex h-7 items-center border-t border-border px-3 text-xs text-muted">
      {text}
    </footer>
  )
}

export { useAppShortcuts }
