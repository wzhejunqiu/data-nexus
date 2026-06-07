import { useMutation, useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { useToastStore } from '@/components/ui/Toast'
import { configApi, type AppConfig } from '@/lib/api/config'
import { formatError } from '@/lib/api/errors'

function SettingsForm({
  initial,
  configPath,
  onClose,
}: {
  initial: AppConfig
  configPath?: string
  onClose: () => void
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const [form, setForm] = useState(initial)

  const save = useMutation({
    mutationFn: () => configApi.updateConfig(form),
    onSuccess: () => {
      pushToast(t('settings.savedRestart'), 'info')
      onClose()
    },
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

  const setLog = (patch: Partial<AppConfig['log']>) =>
    setForm({ ...form, log: { ...form.log, ...patch } })
  const setFile = (patch: Partial<AppConfig['log']['file']>) =>
    setForm({ ...form, log: { ...form.log, file: { ...form.log.file, ...patch } } })

  return (
    <Dialog
      open
      onOpenChange={(next) => !next && onClose()}
      title={t('settings.title')}
      footer={
        <>
          <Button variant="outline" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button onClick={() => save.mutate()} disabled={save.isPending}>
            {t('settings.save')}
          </Button>
        </>
      }
    >
      <div className="space-y-3 text-sm">
        {configPath && (
          <p className="text-xs text-muted">
            {t('settings.configPath')}: <code>{configPath}</code>
          </p>
        )}
        <label className="flex flex-col gap-1">
          {t('settings.logLevel')}
          <select
            className="rounded border border-border bg-transparent px-2 py-1"
            value={form.log.level}
            onChange={(e) => setLog({ level: e.target.value })}
          >
            {['debug', 'info', 'warn', 'error'].map((l) => (
              <option key={l} value={l}>
                {l}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1">
          {t('settings.logOutput')}
          <select
            className="rounded border border-border bg-transparent px-2 py-1"
            value={form.log.output}
            onChange={(e) => setLog({ output: e.target.value })}
          >
            {['auto', 'console', 'file', 'both'].map((o) => (
              <option key={o} value={o}>
                {o}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1">
          {t('settings.logFilePath')}
          <input
            className="rounded border border-border bg-transparent px-2 py-1 font-mono text-xs"
            value={form.log.file.path}
            onChange={(e) => setFile({ path: e.target.value })}
          />
        </label>
        <label className="flex flex-col gap-1">
          {t('settings.maxSizeMB')}
          <input
            type="number"
            className="rounded border border-border bg-transparent px-2 py-1"
            value={form.log.file.max_size_mb}
            onChange={(e) => setFile({ max_size_mb: Number(e.target.value) })}
          />
        </label>
      </div>
    </Dialog>
  )
}

export function SettingsDialog({
  open,
  onOpenChange,
  config,
  configPath,
  isLoading,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  config?: AppConfig
  configPath?: string
  isLoading?: boolean
}) {
  const { t } = useTranslation()

  if (!open) return null

  if (!config) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange} title={t('settings.title')}>
        <p className="text-sm text-muted">{isLoading ? t('common.loading') : '—'}</p>
      </Dialog>
    )
  }

  return (
    <SettingsForm
      key={configPath ?? 'config'}
      initial={config}
      configPath={configPath}
      onClose={() => onOpenChange(false)}
    />
  )
}

export function SettingsDialogContainer({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { data: config, isLoading } = useQuery({
    queryKey: ['config'],
    queryFn: () => configApi.getConfig(),
    enabled: open,
  })
  const { data: configPath } = useQuery({
    queryKey: ['configPath'],
    queryFn: () => configApi.getConfigPath(),
    enabled: open,
  })

  return (
    <SettingsDialog
      open={open}
      onOpenChange={onOpenChange}
      config={config}
      configPath={configPath}
      isLoading={isLoading}
    />
  )
}
