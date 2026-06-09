import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { queryHistoryApi } from '@/lib/api/queryHistory'
import { connectionApi } from '@/lib/api/connection'
import { formatError } from '@/lib/api/errors'
import { Dialog } from '@/components/ui/Dialog'
import { useToastStore } from '@/components/ui/Toast'
import { useWorkspaceStore } from '@/stores/workspaceStore'

export function SqlExecutionHistoryDialog({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (v: boolean) => void
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const pushToast = useToastStore((s) => s.push)
  const setActiveConnectionId = useWorkspaceStore((s) => s.setActiveConnectionId)
  const setActiveTab = useWorkspaceStore((s) => s.setActiveTab)
  const setPendingSql = useWorkspaceStore((s) => s.setPendingSql)

  const { data: history } = useQuery({
    queryKey: ['sql-executions-all'],
    queryFn: () => queryHistoryApi.listAllExecutions(50),
    enabled: open,
  })
  const { data: connections } = useQuery({
    queryKey: ['connections'],
    queryFn: () => connectionApi.list(),
    enabled: open,
  })

  const connName = (id: string) => {
    const conn = connections?.items.find((c) => c.id === id)
    if (conn) return conn.name
    return t('sqlHistory.deletedConnection')
  }

  const finishJump = (connectionId: string, sql: string) => {
    setActiveConnectionId(connectionId)
    setActiveTab('sql')
    setPendingSql(sql)
    onOpenChange(false)
  }

  const jumpTo = async (connectionId: string, sql: string) => {
    const conn = connections?.items.find((c) => c.id === connectionId)
    if (!conn) {
      pushToast(t('sqlHistory.connectionMissing'), 'error')
      return
    }
    if (conn.status === 'open') {
      finishJump(connectionId, sql)
      return
    }
    try {
      await connectionApi.open(connectionId)
      await qc.invalidateQueries({ queryKey: ['connections'] })
      await qc.invalidateQueries({ queryKey: ['connectionSidebarTree'] })
      finishJump(connectionId, sql)
    } catch (err) {
      pushToast(formatError(t, err), 'error')
    }
  }

  const items = history?.items ?? []

  return (
    <Dialog open={open} onOpenChange={onOpenChange} title={t('sqlHistory.title')} wide>
      <div className="max-h-[60vh] overflow-auto">
        {items.length === 0 ? (
          <p className="text-sm text-muted">{t('sqlHistory.empty')}</p>
        ) : (
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-border text-xs text-muted">
                <th className="p-2">{t('sqlHistory.time')}</th>
                <th className="p-2">{t('sqlHistory.connection')}</th>
                <th className="p-2">{t('sqlHistory.sql')}</th>
                <th className="p-2">{t('sqlHistory.kind')}</th>
                <th className="p-2">{t('sqlHistory.duration')}</th>
                <th className="p-2">{t('sqlHistory.effectRows')}</th>
              </tr>
            </thead>
            <tbody>
              {items.map((row, i) => (
                <tr
                  key={row.id ?? i}
                  className="cursor-pointer border-b border-border/50 hover:bg-muted/20"
                  onClick={() => void jumpTo(row.connectionId, row.sql)}
                >
                  <td className="whitespace-nowrap p-2 text-xs">{row.executedAt}</td>
                  <td className="max-w-[8rem] truncate p-2">{connName(row.connectionId)}</td>
                  <td className="max-w-[20rem] truncate p-2 font-mono text-xs">{row.sql}</td>
                  <td className="p-2 text-xs">{row.kind}</td>
                  <td className="p-2 text-xs">{row.durationMs}ms</td>
                  <td className="p-2 text-xs">{row.effectRows}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </Dialog>
  )
}
