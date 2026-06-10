import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { connectionApi } from '@/lib/api/connection'
import { dialogApi } from '@/lib/api/dialog'
import { validateAttachAlias } from '@/lib/attachAlias'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { formatError } from '@/lib/api/errors'
import { useToastStore } from '@/components/ui/Toast'

function suggestAlias(filePath: string): string {
  const base = filePath.split(/[/\\]/).pop() ?? ''
  return base.replace(/\.(db|sqlite|sqlite3)$/i, '')
}

export function AttachDatabaseDialog({
  connectionId,
  open,
  onOpenChange,
  onAttached,
  initialPath,
  closeOnSuccess = true,
  readOnly = false,
}: {
  connectionId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onAttached: () => void
  initialPath?: string
  closeOnSuccess?: boolean
  readOnly?: boolean
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const [filePath, setFilePath] = useState(initialPath ?? '')
  const [alias, setAlias] = useState(initialPath ? suggestAlias(initialPath) : '')
  const [busy, setBusy] = useState(false)

  const { data: attached } = useQuery({
    queryKey: ['attached', connectionId],
    queryFn: () => connectionApi.listAttached(connectionId),
    enabled: open,
  })

  const existingAliases = useMemo(() => {
    const names = ['main']
    for (const item of attached ?? []) {
      names.push(item.alias)
    }
    return names
  }, [attached])

  const effectiveAlias = alias.trim() || (filePath.trim() ? suggestAlias(filePath) : '')
  const aliasError = effectiveAlias
    ? validateAttachAlias(effectiveAlias, existingAliases)
    : filePath.trim()
      ? 'required'
      : null
  const aliasErrorMessage = aliasError ? t(`attach.${aliasError}Alias`) : null

  const handleOpenChange = (next: boolean) => {
    if (!next) {
      setFilePath('')
      setAlias('')
    } else if (initialPath) {
      setFilePath(initialPath)
      setAlias(suggestAlias(initialPath))
    }
    onOpenChange(next)
  }

  const browse = async () => {
    try {
      const path = await dialogApi.openDatabaseFile()
      if (path) {
        setFilePath(path)
        if (!alias.trim()) setAlias(suggestAlias(path))
      }
    } catch {
      /* cancelled */
    }
  }

  const submit = async () => {
    const path = filePath.trim()
    if (!path || aliasError) return
    const attachAlias = effectiveAlias
    setBusy(true)
    try {
      await connectionApi.attach(connectionId, path, attachAlias)
      onAttached()
      if (closeOnSuccess) handleOpenChange(false)
    } catch (err) {
      pushToast(formatError(t, err), 'error')
    } finally {
      setBusy(false)
    }
  }

  const canSubmit = Boolean(filePath.trim()) && !aliasError && !busy

  return (
    <Dialog open={open} onOpenChange={handleOpenChange} title={t('connection.attachDatabase')}>
      <div className="space-y-3 text-sm">
        {readOnly && (
          <p className="text-xs text-amber-600 dark:text-amber-500">{t('attach.readOnlyHint')}</p>
        )}
        <label className="block">
          <span className="text-muted">{t('connection.pathHint')}</span>
          <div className="mt-1 flex gap-2">
            <input
              className="min-w-0 flex-1 rounded border border-border bg-transparent px-2 py-1"
              value={filePath}
              onChange={(e) => setFilePath(e.target.value)}
            />
            <Button type="button" variant="outline" size="sm" onClick={() => void browse()}>
              {t('connection.browse')}
            </Button>
          </div>
        </label>
        <label className="block">
          <span className="text-muted">{t('connection.attachAlias')}</span>
          <input
            className="mt-1 w-full rounded border border-border bg-transparent px-2 py-1"
            value={alias}
            placeholder={t('connection.attachAliasOptional')}
            onChange={(e) => setAlias(e.target.value)}
          />
          {aliasErrorMessage && <p className="mt-1 text-xs text-red-500">{aliasErrorMessage}</p>}
        </label>
        <p className="text-xs text-muted">{t('attach.sessionHint')}</p>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => handleOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button disabled={!canSubmit} onClick={() => void submit()}>
            {t('connection.attach')}
          </Button>
        </div>
      </div>
    </Dialog>
  )
}
