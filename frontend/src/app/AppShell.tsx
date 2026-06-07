import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { Header, StatusBar } from '@/components/Header'
import { Tabs } from '@/components/ui/Tabs'
import { Toaster, useToastStore } from '@/components/ui/Toast'
import { ConnectionTree } from '@/features/connection/ConnectionTree'
import { SchemaTable } from '@/features/schema/SchemaTable'
import { SqlEditor } from '@/features/sql-editor/SqlEditor'
import { DataGrid } from '@/features/table-browser/DataGrid'
import { appApi } from '@/lib/api/app'
import { connectionApi } from '@/lib/api/connection'
import { useStatusStore } from '@/stores/statusStore'
import { resolveTheme, useThemeStore } from '@/stores/themeStore'
import { useWorkspaceStore } from '@/stores/workspaceStore'

export function AppShell() {
  const { t, i18n } = useTranslation()
  const qc = useQueryClient()
  const activeTab = useWorkspaceStore((s) => s.activeTab)
  const setActiveTab = useWorkspaceStore((s) => s.setActiveTab)
  const activeConnectionId = useWorkspaceStore((s) => s.activeConnectionId)
  const setActiveConnectionId = useWorkspaceStore((s) => s.setActiveConnectionId)
  const selectedTable = useWorkspaceStore((s) => s.selectedTable)
  const rowCount = useStatusStore((s) => s.rowCount)
  const durationMs = useStatusStore((s) => s.durationMs)

  const { data: connections } = useQuery({
    queryKey: ['connections'],
    queryFn: () => connectionApi.list(),
  })

  const { data: version } = useQuery({
    queryKey: ['version'],
    queryFn: () => appApi.getVersion(),
  })

  const openConnections = connections?.items.filter((c) => c.status === 'open') ?? []
  const openCount = openConnections.length
  const activeConn = openConnections.find((c) => c.id === activeConnectionId)

  useEffect(() => {
    const title = activeConn?.name ? `Data Nexus — ${activeConn.name}` : 'Data Nexus'
    void appApi.setWindowTitle(title)
  }, [activeConn?.name])

  useEffect(() => {
    void appApi.setActiveConnection(activeConnectionId ?? '')
  }, [activeConnectionId])

  useEffect(() => {
    const pushToast = useToastStore.getState().push
    const unsubs = [
      EventsOn('app:connections-changed', () => {
        qc.invalidateQueries({ queryKey: ['connections'] })
      }),
      EventsOn('app:toast', (payload: { message?: string; variant?: 'error' | 'info' }) => {
        if (payload?.message) {
          pushToast(payload.message, payload.variant ?? 'error')
        }
      }),
      EventsOn('app:theme', (mode: string) => {
        if (mode === 'light' || mode === 'dark' || mode === 'system') {
          useThemeStore.getState().setMode(mode as 'light' | 'dark' | 'system')
          document.documentElement.classList.toggle(
            'dark',
            resolveTheme(mode as 'light' | 'dark' | 'system') === 'dark',
          )
        }
      }),
      EventsOn('app:language', (lang: string) => {
        i18n.changeLanguage(lang)
        localStorage.setItem('data-nexus-lang', lang)
      }),
    ]
    return () => unsubs.forEach((u) => u())
  }, [i18n, qc])

  const tabs = [
    { id: 'schema', label: t('tabs.schema') },
    { id: 'data', label: t('tabs.data') },
    { id: 'sql', label: t('tabs.sql') },
  ]

  const statusParts = [
    rowCount != null ? t('status.rows', { count: rowCount }) : null,
    durationMs != null ? t('status.duration', { ms: durationMs }) : null,
    `v${version?.version ?? '0.1.0'}`,
  ].filter(Boolean)

  return (
    <div className="flex h-screen flex-col bg-background text-foreground">
      <Header openCount={openCount} />
      <div className="flex min-h-0 flex-1">
        <ConnectionTree />
        <main className="flex min-w-0 flex-1 flex-col">
          <div className="flex items-center justify-between border-b border-border pr-3">
            <Tabs
              tabs={tabs}
              active={activeTab}
              onChange={(id) => setActiveTab(id as typeof activeTab)}
            />
            <label className="flex items-center gap-2 text-sm text-muted">
              {t('tabs.activeConnection')}
              <select
                className="rounded border border-border bg-transparent px-2 py-1 text-sm text-foreground"
                value={activeConnectionId ?? ''}
                onChange={(e) => setActiveConnectionId(e.target.value || null)}
              >
                <option value="">—</option>
                {openConnections.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
          </div>
          <div className="min-h-0 flex-1 overflow-auto">
            {!activeConnectionId || !selectedTable ? (
              activeTab !== 'sql' ? (
                <p className="p-6 text-sm text-muted">{t('data.selectTable')}</p>
              ) : (
                <SqlEditor connectionId={activeConnectionId} />
              )
            ) : (
              <>
                {activeTab === 'schema' && (
                  <SchemaTable connectionId={activeConnectionId} tableName={selectedTable} />
                )}
                {activeTab === 'data' && (
                  <DataGrid connectionId={activeConnectionId} tableName={selectedTable} />
                )}
                {activeTab === 'sql' && <SqlEditor connectionId={activeConnectionId} />}
              </>
            )}
          </div>
        </main>
      </div>
      <StatusBar text={statusParts.join(' · ')} />
      <Toaster />
    </div>
  )
}
