import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { AboutDialog } from '@/components/AboutDialog'
import { Header, StatusBar, useAppShortcuts } from '@/components/Header'
import { Tabs } from '@/components/ui/Tabs'
import { Toaster, useToastStore } from '@/components/ui/Toast'
import { ConnectionTree } from '@/features/connection/ConnectionTree'
import { NewConnectionDialog } from '@/features/connection/NewConnectionDialog'
import { ExportWizardPage } from '@/features/csv/ExportWizardPage'
import { ImportWizardPage } from '@/features/csv/ImportWizardPage'
import { SchemaTable } from '@/features/schema/SchemaTable'
import { SqlEditor } from '@/features/sql-editor/SqlEditor'
import { SettingsDialogContainer } from '@/features/settings/SettingsDialog'
import { SqlExecutionHistoryDialog } from '@/features/sql-history/SqlExecutionHistoryDialog'
import { DataGrid } from '@/features/table-browser/DataGrid'
import { appApi } from '@/lib/api/app'
import { connectionApi } from '@/lib/api/connection'
import { connectionGroupApi } from '@/lib/api/connectionGroup'
import { dialogApi } from '@/lib/api/dialog'
import { formatError } from '@/lib/api/errors'
import { useExportStore } from '@/stores/exportStore'
import { useImportStore } from '@/stores/importStore'
import { setAppLanguage, isAppLanguage } from '@/i18n/language'
import { useStatusStore } from '@/stores/statusStore'
import { applyThemeMode, isThemeMode } from '@/stores/themeStore'
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
  const exportSession = useExportStore((s) => s.session)
  const importSession = useImportStore((s) => s.session)

  const [newConnectionOpen, setNewConnectionOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [sqlHistoryOpen, setSqlHistoryOpen] = useState(false)
  const [aboutOpen, setAboutOpen] = useState(false)

  const wizardActive = Boolean(exportSession || importSession)

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

  const createRootGroup = useCallback(async () => {
    try {
      await connectionGroupApi.createGroup('', t('connectionGroup.newGroupName'))
      qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
      qc.invalidateQueries({ queryKey: ['connections'] })
    } catch (err) {
      useToastStore.getState().push(formatError(t, err), 'error')
    }
  }, [qc, t])

  const menuBarProps = {
    openCount,
    wizardTitle: exportSession
      ? t('csv.exportWizardTitle')
      : importSession
        ? t('csv.importWizardTitle')
        : undefined,
    activeConnectionId,
    onNewConnection: () => setNewConnectionOpen(true),
    onNewGroup: () => void createRootGroup(),
    onOpenSettings: () => setSettingsOpen(true),
    onOpenSqlHistory: () => setSqlHistoryOpen(true),
    onOpenAbout: () => setAboutOpen(true),
    disabled: wizardActive,
  }

  const { handleKeyDown } = useAppShortcuts(menuBarProps)

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  useEffect(() => {
    const title = activeConn?.name ? `Data Nexus — ${activeConn.name}` : 'Data Nexus'
    void appApi.setWindowTitle(title)
  }, [activeConn?.name])

  useEffect(() => {
    void appApi.setActiveConnection(activeConnectionId ?? '')
  }, [activeConnectionId])

  useEffect(() => {
    const pushToast = useToastStore.getState().push
    const invalidateConnections = () => {
      qc.invalidateQueries({ queryKey: ['connections'] })
      qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
    }
    const unsubs = [
      EventsOn('app:connections-changed', invalidateConnections),
      EventsOn('app:connection-opened', (payload: { id?: string }) => {
        if (payload?.id) setActiveConnectionId(payload.id)
      }),
      EventsOn('app:new-connection', () => setNewConnectionOpen(true)),
      EventsOn('app:new-group', () => {
        void createRootGroup()
      }),
      EventsOn('app:open-sqlite', () => {
        void (async () => {
          try {
            const path = await dialogApi.openDatabaseFile()
            if (path) {
              const conn = await connectionApi.openFromFile({
                filePath: path,
                readOnly: false,
                wal: false,
              })
              setActiveConnectionId(conn.id)
              invalidateConnections()
            }
          } catch {
            /* cancelled or error via toast */
          }
        })()
      }),
      EventsOn('app:close-connection', () => {
        void (async () => {
          const id = useWorkspaceStore.getState().activeConnectionId
          if (!id) return
          await connectionApi.close(id)
          setActiveConnectionId(null)
          useWorkspaceStore.getState().setSelectedTable(null)
          invalidateConnections()
        })()
      }),
      EventsOn('app:sql-history', () => setSqlHistoryOpen(true)),
      EventsOn('app:about', () => setAboutOpen(true)),
      EventsOn('app:settings', () => setSettingsOpen(true)),
      EventsOn('app:toast', (payload: { message?: string; variant?: 'error' | 'info' }) => {
        if (payload?.message) pushToast(payload.message, payload.variant ?? 'error')
      }),
      EventsOn('app:theme', (mode: string) => {
        if (isThemeMode(mode)) applyThemeMode(mode)
      }),
      EventsOn('app:language', (lang: string) => {
        if (isAppLanguage(lang)) setAppLanguage(lang)
      }),
    ]
    return () => unsubs.forEach((u) => u())
  }, [createRootGroup, i18n, qc, setActiveConnectionId])

  const tabs = [
    { id: 'schema', label: t('tabs.schema') },
    { id: 'data', label: t('tabs.data') },
    { id: 'sql', label: t('tabs.sql') },
  ]

  const statusParts = [
    rowCount != null ? t('status.rows', { count: rowCount }) : null,
    durationMs != null ? t('status.duration', { ms: durationMs }) : null,
    `v${version?.version ?? '0.5.0'}`,
  ].filter(Boolean)

  return (
    <div className="flex h-screen flex-col bg-background text-foreground">
      <Header {...menuBarProps} />
      {exportSession ? (
        <ExportWizardPage />
      ) : importSession ? (
        <ImportWizardPage />
      ) : (
        <>
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
                      <DataGrid
                        key={`${activeConnectionId}-${selectedTable}`}
                        connectionId={activeConnectionId}
                        tableName={selectedTable}
                      />
                    )}
                    {activeTab === 'sql' && <SqlEditor connectionId={activeConnectionId} />}
                  </>
                )}
              </div>
            </main>
          </div>
          <StatusBar text={statusParts.join(' · ')} />
        </>
      )}
      <NewConnectionDialog
        open={newConnectionOpen}
        onOpenChange={setNewConnectionOpen}
        readOnly={false}
        wal={false}
      />
      <SettingsDialogContainer open={settingsOpen} onOpenChange={setSettingsOpen} />
      <SqlExecutionHistoryDialog open={sqlHistoryOpen} onOpenChange={setSqlHistoryOpen} />
      <AboutDialog open={aboutOpen} onOpenChange={setAboutOpen} />
      <Toaster />
    </div>
  )
}
