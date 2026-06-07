import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { queriesApi } from '@/lib/api/queries'
import { formatError } from '@/lib/api/errors'
import { useWorkspaceStore } from '@/stores/workspaceStore'
import { useToastStore } from '@/components/ui/Toast'

export function SavedQueries() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const pushToast = useToastStore((s) => s.push)
  const setActiveTab = useWorkspaceStore((s) => s.setActiveTab)
  const setPendingSql = useWorkspaceStore((s) => s.setPendingSql)
  const activeConnectionId = useWorkspaceStore((s) => s.activeConnectionId)
  const [saveOpen, setSaveOpen] = useState(false)
  const [queryName, setQueryName] = useState('')
  const [saving, setSaving] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: ['canned-queries'],
    queryFn: () => queriesApi.list(),
  })

  const remove = useMutation({
    mutationFn: (id: string) => queriesApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['canned-queries'] }),
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

  const items = data?.items ?? []

  const selectQuery = (sql: string) => {
    setPendingSql(sql)
    setActiveTab('sql')
  }

  const confirmSave = async () => {
    const name = queryName.trim()
    if (!name) return
    const sql = useWorkspaceStore.getState().editorSql
    setSaving(true)
    try {
      await queriesApi.save({
        name,
        sql: sql || 'SELECT 1;',
        connectionId: activeConnectionId,
      })
      await qc.invalidateQueries({ queryKey: ['canned-queries'] })
      pushToast(t('savedQueries.saved'), 'info')
      setSaveOpen(false)
      setQueryName('')
    } catch (err) {
      pushToast(formatError(t, err), 'error')
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <div className="border-b border-border p-3">
        <div className="mb-2 flex items-center justify-between gap-2">
          <p className="text-xs font-semibold text-muted">{t('savedQueries.title')}</p>
          <Button
            size="sm"
            variant="outline"
            onClick={() => {
              setQueryName('')
              setSaveOpen(true)
            }}
          >
            {t('savedQueries.save')}
          </Button>
        </div>
        {isLoading && <p className="text-xs text-muted">{t('common.loading')}</p>}
        {items.length === 0 && !isLoading && (
          <p className="text-xs text-muted">{t('savedQueries.empty')}</p>
        )}
        <ul className="max-h-32 space-y-0.5 overflow-y-auto">
          {items.map((q) => (
            <li key={q.id} className="flex items-center gap-1">
              <button
                type="button"
                className="flex-1 truncate rounded px-2 py-1 text-left text-xs hover:bg-muted/40"
                onClick={() => selectQuery(q.sql)}
                title={q.sql}
              >
                {q.name}
              </button>
              <button
                type="button"
                className="px-1 text-xs text-muted hover:text-red-500"
                onClick={() => remove.mutate(q.id)}
              >
                ×
              </button>
            </li>
          ))}
        </ul>
      </div>
      <Dialog
        open={saveOpen}
        onOpenChange={setSaveOpen}
        title={t('savedQueries.namePrompt')}
        footer={
          <>
            <Button variant="outline" onClick={() => setSaveOpen(false)} disabled={saving}>
              {t('common.cancel')}
            </Button>
            <Button onClick={() => void confirmSave()} disabled={saving || !queryName.trim()}>
              {t('common.confirm')}
            </Button>
          </>
        }
      >
        <input
          autoFocus
          className="w-full rounded border border-border bg-transparent px-3 py-2 text-sm"
          placeholder={t('savedQueries.namePrompt')}
          value={queryName}
          onChange={(e) => setQueryName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && queryName.trim()) void confirmSave()
          }}
        />
      </Dialog>
    </>
  )
}
