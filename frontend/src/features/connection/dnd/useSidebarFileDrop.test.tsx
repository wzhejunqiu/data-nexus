import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { cleanup, render, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useSidebarFileDrop } from './useSidebarFileDrop'

const openSqliteAndMaybePlace = vi.fn()
const pushToast = vi.fn()
let fileDropHandler: ((payload: { x?: number; y?: number; paths?: string[] }) => void) | null = null

vi.mock('../placeConnectionInGroup', () => ({
  openSqliteAndMaybePlace: (...args: unknown[]) => openSqliteAndMaybePlace(...args),
}))

vi.mock('@/components/ui/Toast', () => ({
  useToastStore: (selector: (s: { push: typeof pushToast }) => unknown) =>
    selector({ push: pushToast }),
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

vi.mock('../../../../wailsjs/runtime/runtime', () => ({
  EventsOn: (_event: string, handler: typeof fileDropHandler) => {
    fileDropHandler = handler
    return () => {
      fileDropHandler = null
    }
  },
}))

function DropHarness({
  onAttach,
  onRefresh,
}: {
  onAttach: (target: { connectionId: string; paths: string[] }) => void
  onRefresh: () => void
}) {
  useSidebarFileDrop({ onAttach, onRefresh })
  return (
    <div>
      <div data-group-drop-target="" data-group-id="g1">
        Group
      </div>
      <div
        data-sqlite-drop-target=""
        data-connection-id="c-open"
        data-connection-type="sqlite"
        data-connection-open="true"
      >
        Open SQLite
      </div>
      <div data-testid="blank">Blank</div>
    </div>
  )
}

function mockElementFromPoint(el: HTMLElement | null) {
  document.elementFromPoint = vi.fn(() => el) as typeof document.elementFromPoint
}

describe('useSidebarFileDrop', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    fileDropHandler = null
    ;(globalThis as typeof globalThis & { runtime?: { EventsOnMultiple?: unknown } }).runtime = {
      EventsOnMultiple: true,
    }
    openSqliteAndMaybePlace.mockResolvedValue({ id: 'new-conn' })
  })

  afterEach(() => {
    cleanup()
    const g = globalThis as typeof globalThis & { runtime?: { EventsOnMultiple?: unknown } }
    g.runtime = undefined
    Reflect.deleteProperty(document, 'elementFromPoint')
  })

  it('opens to group when dropped on group target', async () => {
    const onAttach = vi.fn()
    const onRefresh = vi.fn()
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={qc}>
        <DropHarness onAttach={onAttach} onRefresh={onRefresh} />
      </QueryClientProvider>,
    )

    const groupEl = document.querySelector('[data-group-id="g1"]') as HTMLElement
    mockElementFromPoint(groupEl)

    fileDropHandler?.({ x: 10, y: 10, paths: ['/tmp/a.db'] })

    await waitFor(() => {
      expect(openSqliteAndMaybePlace).toHaveBeenCalledWith(
        { filePath: '/tmp/a.db', readOnly: false, wal: false },
        'g1',
        expect.any(QueryClient),
      )
      expect(onRefresh).toHaveBeenCalled()
    })
    expect(onAttach).not.toHaveBeenCalled()
  })

  it('opens free when dropped on blank area', async () => {
    const onAttach = vi.fn()
    const onRefresh = vi.fn()
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={qc}>
        <DropHarness onAttach={onAttach} onRefresh={onRefresh} />
      </QueryClientProvider>,
    )

    const blankEl = document.querySelector('[data-testid="blank"]') as HTMLElement
    mockElementFromPoint(blankEl)

    fileDropHandler?.({ x: 5, y: 5, paths: ['/tmp/free.db'] })

    await waitFor(() => {
      expect(openSqliteAndMaybePlace).toHaveBeenCalledWith(
        { filePath: '/tmp/free.db', readOnly: false, wal: false },
        null,
        expect.any(QueryClient),
      )
    })
  })

  it('attaches when dropped on open sqlite connection', async () => {
    const onAttach = vi.fn()
    const onRefresh = vi.fn()
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={qc}>
        <DropHarness onAttach={onAttach} onRefresh={onRefresh} />
      </QueryClientProvider>,
    )

    const sqliteEl = document.querySelector('[data-connection-id="c-open"]') as HTMLElement
    mockElementFromPoint(sqliteEl)

    fileDropHandler?.({ x: 20, y: 20, paths: ['/tmp/attach.db'] })

    expect(onAttach).toHaveBeenCalledWith({
      connectionId: 'c-open',
      paths: ['/tmp/attach.db'],
    })
    expect(openSqliteAndMaybePlace).not.toHaveBeenCalled()
  })
})
