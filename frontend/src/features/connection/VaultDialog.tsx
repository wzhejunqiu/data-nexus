import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'
import { useToastStore } from '@/components/ui/Toast'
import { secretsApi } from '@/lib/api/secrets'
import { formatError } from '@/lib/api/errors'
import { isVaultWrongPasswordError } from '@/lib/api/secrets'

export type VaultDialogMode = 'init' | 'unlock'

export function VaultDialog({
  open,
  mode,
  onOpenChange,
  onSuccess,
}: {
  open: boolean
  mode: VaultDialogMode
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}) {
  const { t } = useTranslation()
  const pushToast = useToastStore((s) => s.push)
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [busy, setBusy] = useState(false)

  const title = mode === 'init' ? t('vault.initTitle') : t('vault.unlockTitle')
  const canSubmit =
    mode === 'init' ? password.length >= 8 && password === confirm : password.length >= 8

  const submit = async () => {
    setBusy(true)
    try {
      if (mode === 'init') {
        await secretsApi.initVault(password)
        pushToast(t('vault.initSuccess'), 'success')
      } else {
        await secretsApi.unlockVault(password)
        pushToast(t('vault.unlockSuccess'), 'success')
      }
      setPassword('')
      setConfirm('')
      onOpenChange(false)
      onSuccess()
    } catch (err) {
      if (isVaultWrongPasswordError(err)) {
        pushToast(t('vault.wrongPassword'), 'error')
      } else {
        pushToast(formatError(t, err), 'error')
      }
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={title}
      footer={
        <>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={() => void submit()} disabled={!canSubmit || busy}>
            {mode === 'init' ? t('vault.initAction') : t('vault.unlockAction')}
          </Button>
        </>
      }
    >
      <div className="space-y-3">
        <p className="text-xs text-muted">
          {mode === 'init' ? t('vault.initHint') : t('vault.unlockHint')}
        </p>
        <Input
          type="password"
          placeholder={t('vault.masterPassword')}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        {mode === 'init' && (
          <Input
            type="password"
            placeholder={t('vault.confirmPassword')}
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
          />
        )}
      </div>
    </Dialog>
  )
}
