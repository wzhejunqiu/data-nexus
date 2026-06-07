import { useMutation, useQuery } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { useToastStore } from '@/components/ui/Toast'
import { configApi, type AppConfig } from '@/lib/api/config'
import { formatError } from '@/lib/api/errors'
import { APP_LANGUAGES, getAppLanguage, setAppLanguage, type AppLanguage } from '@/i18n/language'
import {
  THEME_MODES,
  applyThemeMode,
  isThemeMode,
  useThemeStore,
  type ThemeMode,
} from '@/stores/themeStore'

const LOG_LEVELS = ['debug', 'info', 'warn', 'error'] as const
const LOG_OUTPUTS = ['auto', 'console', 'file', 'both'] as const

function validateSettingsForm(form: AppConfig): Record<string, string> {
  const errors: Record<string, string> = {}
  if (!LOG_LEVELS.includes(form.log.level as (typeof LOG_LEVELS)[number])) {
    errors.level = 'invalid'
  }
  if (!LOG_OUTPUTS.includes(form.log.output as (typeof LOG_OUTPUTS)[number])) {
    errors.output = 'invalid'
  }
  if (form.log.file.max_size_mb < 1) errors.max_size_mb = 'invalid'
  if (form.log.file.max_backups < 1) errors.max_backups = 'invalid'
  if (form.log.file.max_age_days < 1) errors.max_age_days = 'invalid'
  if (form.log.file.path && form.log.file.path.includes('..')) {
    errors.path = 'invalid'
  }
  return errors
}

function needsRestartForSink(initial: AppConfig, next: AppConfig): boolean {
  return (
    initial.log.output !== next.log.output ||
    initial.log.file.path !== next.log.file.path ||
    initial.log.file.max_size_mb !== next.log.file.max_size_mb ||
    initial.log.file.max_backups !== next.log.file.max_backups ||
    initial.log.file.max_age_days !== next.log.file.max_age_days ||
    initial.log.file.compress !== next.log.file.compress
  )
}

function FieldError({ message }: { message?: string }) {
  if (!message) return null
  return <span className="text-xs text-red-500">{message}</span>
}

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
  const [touched, setTouched] = useState(false)
  const [lang, setLang] = useState<AppLanguage>(() => getAppLanguage())
  const themeMode = useThemeStore((s) => s.mode)

  const errors = useMemo(() => validateSettingsForm(form), [form])
  const hasErrors = Object.keys(errors).length > 0

  const save = useMutation({
    mutationFn: () => configApi.updateConfig(form),
    onSuccess: () => {
      const restart = needsRestartForSink(initial, form)
      pushToast(t(restart ? 'settings.savedRestartRequired' : 'settings.savedLevelApplied'), 'info')
      onClose()
    },
    onError: (err) => pushToast(formatError(t, err), 'error'),
  })

  const setLog = (patch: Partial<AppConfig['log']>) =>
    setForm({ ...form, log: { ...form.log, ...patch } })
  const setFile = (patch: Partial<AppConfig['log']['file']>) =>
    setForm({ ...form, log: { ...form.log, file: { ...form.log.file, ...patch } } })

  const errMsg = (key: string) =>
    touched && errors[key] ? t('settings.validationInvalid') : undefined

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
          <Button
            onClick={() => {
              setTouched(true)
              if (!hasErrors) save.mutate()
            }}
            disabled={save.isPending}
          >
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
          {t('settings.language')}
          <select
            className="rounded border border-border bg-transparent px-2 py-1"
            value={lang}
            onChange={(e) => {
              const next = e.target.value
              if (!APP_LANGUAGES.includes(next as AppLanguage)) return
              const value = next as AppLanguage
              setLang(value)
              setAppLanguage(value)
            }}
          >
            {APP_LANGUAGES.map((code) => (
              <option key={code} value={code}>
                {code === 'zh-CN' ? t('language.zh') : t('language.en')}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1">
          {t('settings.theme')}
          <select
            className="rounded border border-border bg-transparent px-2 py-1"
            value={themeMode}
            onChange={(e) => {
              const next = e.target.value
              if (!isThemeMode(next)) return
              applyThemeMode(next as ThemeMode)
            }}
          >
            {THEME_MODES.map((mode) => (
              <option key={mode} value={mode}>
                {t(`theme.${mode}`)}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1">
          {t('settings.logLevel')}
          <select
            className="rounded border border-border bg-transparent px-2 py-1"
            value={form.log.level}
            onChange={(e) => setLog({ level: e.target.value })}
          >
            {LOG_LEVELS.map((l) => (
              <option key={l} value={l}>
                {l}
              </option>
            ))}
          </select>
          <FieldError message={errMsg('level')} />
        </label>
        <label className="flex flex-col gap-1">
          {t('settings.logOutput')}
          <select
            className="rounded border border-border bg-transparent px-2 py-1"
            value={form.log.output}
            onChange={(e) => setLog({ output: e.target.value })}
          >
            {LOG_OUTPUTS.map((o) => (
              <option key={o} value={o}>
                {o}
              </option>
            ))}
          </select>
          <FieldError message={errMsg('output')} />
        </label>
        <label className="flex flex-col gap-1">
          {t('settings.logFilePath')}
          <input
            className="rounded border border-border bg-transparent px-2 py-1 font-mono text-xs"
            value={form.log.file.path}
            onChange={(e) => setFile({ path: e.target.value })}
          />
          <FieldError message={errMsg('path')} />
        </label>
        <label className="flex flex-col gap-1">
          {t('settings.maxSizeMB')}
          <input
            type="number"
            min={1}
            className="rounded border border-border bg-transparent px-2 py-1"
            value={form.log.file.max_size_mb}
            onChange={(e) => setFile({ max_size_mb: Number(e.target.value) })}
          />
          <FieldError message={errMsg('max_size_mb')} />
        </label>
        <label className="flex flex-col gap-1">
          {t('settings.maxBackups')}
          <input
            type="number"
            min={1}
            className="rounded border border-border bg-transparent px-2 py-1"
            value={form.log.file.max_backups}
            onChange={(e) => setFile({ max_backups: Number(e.target.value) })}
          />
          <FieldError message={errMsg('max_backups')} />
        </label>
        <label className="flex flex-col gap-1">
          {t('settings.maxAgeDays')}
          <input
            type="number"
            min={1}
            className="rounded border border-border bg-transparent px-2 py-1"
            value={form.log.file.max_age_days}
            onChange={(e) => setFile({ max_age_days: Number(e.target.value) })}
          />
          <FieldError message={errMsg('max_age_days')} />
        </label>
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={form.log.file.compress}
            onChange={(e) => setFile({ compress: e.target.checked })}
          />
          {t('settings.compress')}
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
