import { describe, expect, it, vi, beforeEach } from 'vitest'
import { QueryClient } from '@tanstack/react-query'
import { closeActiveConnection, invalidateConnectionQueries } from './connectionLifecycle'
import { openSqliteAndMaybePlace } from './placeConnectionInGroup'
import { useWorkspaceStore } from '@/stores/workspaceStore'

vi.mock('@/lib/api/connection', () => ({
  connectionApi: {
    close: vi.fn(),
    openFromFile: vi.fn(),
  },
}))

vi.mock('@/lib/api/connectionGroup', () => ({
  connectionGroupApi: {
    moveConnectionToGroup: vi.fn(),
  },
}))

describe('connectionLifecycle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useWorkspaceStore.setState({
      activeConnectionId: 'conn-1',
      selectedTable: 'main.users',
    })
  })

  it('invalidateConnectionQueries refreshes connections and sidebar tree', () => {
    const qc = new QueryClient()
    const spy = vi.spyOn(qc, 'invalidateQueries')
    invalidateConnectionQueries(qc)
    expect(spy).toHaveBeenCalledWith({ queryKey: ['connections'] })
    expect(spy).toHaveBeenCalledWith({ queryKey: ['connectionSidebarTree'] })
  })

  it('closeActiveConnection clears workspace and invalidates', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    vi.mocked(connectionApi.close).mockResolvedValue(undefined)
    const qc = new QueryClient()
    const spy = vi.spyOn(qc, 'invalidateQueries')

    await closeActiveConnection(qc, 'conn-1')

    expect(connectionApi.close).toHaveBeenCalledWith('conn-1')
    expect(useWorkspaceStore.getState().activeConnectionId).toBeNull()
    expect(useWorkspaceStore.getState().selectedTable).toBeNull()
    expect(spy).toHaveBeenCalledWith({ queryKey: ['connections'] })
    expect(spy).toHaveBeenCalledWith({ queryKey: ['connectionSidebarTree'] })
  })

  it('openSqliteAndMaybePlace invalidates sidebar even without group', async () => {
    const { connectionApi } = await import('@/lib/api/connection')
    const { connectionGroupApi } = await import('@/lib/api/connectionGroup')
    vi.mocked(connectionApi.openFromFile).mockResolvedValue({
      id: 'new-1',
      displayName: 'app.db',
      type: 'sqlite',
      config: { type: 'sqlite', sqlite: { filePath: '/tmp/app.db', readOnly: false } },
      connectedAt: '2026-01-01T00:00:00Z',
    })
    const qc = new QueryClient()
    const spy = vi.spyOn(qc, 'invalidateQueries')

    await openSqliteAndMaybePlace(
      { filePath: '/tmp/app.db', readOnly: false, wal: false },
      null,
      qc,
    )

    expect(connectionGroupApi.moveConnectionToGroup).not.toHaveBeenCalled()
    expect(spy).toHaveBeenCalledWith({ queryKey: ['connections'] })
    expect(spy).toHaveBeenCalledWith({ queryKey: ['connectionSidebarTree'] })
  })
})
