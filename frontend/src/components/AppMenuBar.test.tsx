import { describe, expect, it, vi, beforeEach } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AppMenuBar, useAppShortcuts } from './AppMenuBar'
import { useWorkspaceStore } from '@/stores/workspaceStore'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}))

vi.mock('../../wailsjs/runtime/runtime', () => ({
  Quit: vi.fn(),
}))

vi.mock('@/lib/platform', () => ({
  useIsMacOS: () => false,
}))

vi.mock('@/lib/api/dialog', () => ({
  dialogApi: {
    openDatabaseFile: vi.fn().mockResolvedValue('/tmp/app.db'),
  },
}))

vi.mock('@/features/connection/connectionLifecycle', () => ({
  closeActiveConnection: vi.fn(),
}))

vi.mock('@/features/connection/placeConnectionInGroup', () => ({
  openSqliteAndMaybePlace: vi.fn().mockResolvedValue({ id: 'c1' }),
}))

const menuProps = {
  openCount: 1,
  activeConnectionId: 'conn-1',
  onNewConnection: vi.fn(),
  onNewGroup: vi.fn(),
  onOpenSettings: vi.fn(),
  onOpenSqlHistory: vi.fn(),
  onOpenAbout: vi.fn(),
}

function renderMenuBar() {
  const qc = new QueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <AppMenuBar {...menuProps} />
    </QueryClientProvider>,
  )
}

describe('AppMenuBar', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useWorkspaceStore.setState({ activeConnectionId: 'conn-1', selectedGroupId: null })
  })

  it('calls closeActiveConnection from file menu', async () => {
    const { closeActiveConnection } = await import('@/features/connection/connectionLifecycle')
    renderMenuBar()
    const fileTrigger = screen.getByText('menu.file.label')
    fireEvent.pointerDown(fileTrigger, { button: 0, pointerType: 'mouse' })
    fireEvent.click(await screen.findByText('menu.file.closeConnection'))
    expect(closeActiveConnection).toHaveBeenCalledWith(expect.any(QueryClient), 'conn-1')
  })

  it('useAppShortcuts Ctrl+W calls closeActiveConnection', async () => {
    const { closeActiveConnection } = await import('@/features/connection/connectionLifecycle')
    const qc = new QueryClient()
    let handler: (e: KeyboardEvent) => void = () => {}
    function Harness() {
      const { handleKeyDown } = useAppShortcuts(menuProps)
      handler = handleKeyDown
      return null
    }
    render(
      <QueryClientProvider client={qc}>
        <Harness />
      </QueryClientProvider>,
    )
    const event = new KeyboardEvent('keydown', { key: 'w', ctrlKey: true })
    const prevent = vi.spyOn(event, 'preventDefault')
    handler(event)
    expect(prevent).toHaveBeenCalled()
    expect(closeActiveConnection).toHaveBeenCalledWith(qc, 'conn-1')
  })
})
