import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Header, StatusBar } from '@/components/Header'
import { Tabs } from '@/components/ui/Tabs'
import { ConnectionTree } from '@/features/connection/ConnectionTree'
import { SchemaTable } from '@/features/schema/SchemaTable'
import { SqlEditor } from '@/features/sql-editor/SqlEditor'
import { DataGrid } from '@/features/table-browser/DataGrid'
import { appApi } from '@/lib/api/app'
import { connectionApi } from '@/lib/api/connection'
import { useWorkspaceStore } from '@/stores/workspaceStore'

export function AppShell() {
  const { t } = useTranslation()
  const activeTab = useWorkspaceStore((s) => s.activeTab)
  const setActiveTab = useWorkspaceStore((s) => s.setActiveTab)
  const activeConnectionId = useWorkspaceStore((s) => s.activeConnectionId)
  const selectedTable = useWorkspaceStore((s) => s.selectedTable)

  const { data: connections } = useQuery({
    queryKey: ['connections'],
    queryFn: () => connectionApi.list(),
  })

  const { data: version } = useQuery({
    queryKey: ['version'],
    queryFn: () => appApi.getVersion(),
  })

  const openCount = connections?.items.filter((c) => c.status === 'open').length ?? 0

  const tabs = [
    { id: 'schema', label: t('tabs.schema') },
    { id: 'data', label: t('tabs.data') },
    { id: 'sql', label: t('tabs.sql') },
  ]

  return (
    <div className="flex h-screen flex-col bg-background text-foreground">
      <Header openCount={openCount} />
      <div className="flex min-h-0 flex-1">
        <ConnectionTree />
        <main className="flex min-w-0 flex-1 flex-col">
          <Tabs
            tabs={tabs}
            active={activeTab}
            onChange={(id) => setActiveTab(id as typeof activeTab)}
          />
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
      <StatusBar
        text={`data-nexus v${version?.version ?? '0.1.0'} · ${version?.platform ?? ''}/${version?.arch ?? ''}`}
      />
    </div>
  )
}
