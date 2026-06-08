import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { fieldErrorClass, FormField } from '../FormField'
import type { ConnectionFormErrors } from '../connectionFormValidation'
import type { ConnectionFormState } from '../connectionFormDefaults'

export function SQLiteFields({
  state,
  errors,
  mode,
  onChange,
  onBrowse,
}: {
  state: ConnectionFormState
  errors: ConnectionFormErrors
  mode: 'create' | 'edit'
  onChange: (patch: Partial<ConnectionFormState>) => void
  onBrowse?: () => Promise<string | undefined>
}) {
  const { t } = useTranslation()
  const optionalLabel = t('connectionForm.optional')

  return (
    <div className="space-y-3">
      <FormField
        id="displayName"
        label={t('connectionForm.fields.displayName')}
        optional={mode === 'create' ? optionalLabel : undefined}
        error={
          errors.displayName
            ? t('connectionForm.validation.required', {
                field: t('connectionForm.fields.displayName'),
              })
            : undefined
        }
      >
        <Input
          id="displayName"
          value={state.displayName}
          aria-required={false}
          aria-invalid={!!errors.displayName}
          className={fieldErrorClass(!!errors.displayName)}
          onChange={(e) => onChange({ displayName: e.target.value })}
        />
      </FormField>
      <FormField
        id="filePath"
        label={t('connectionForm.fields.filePath')}
        required
        error={
          errors.filePath
            ? t('connectionForm.validation.required', {
                field: t('connectionForm.fields.filePath'),
              })
            : undefined
        }
      >
        {mode === 'edit' ? (
          <Input id="filePath" value={state.filePath} readOnly className="bg-muted/20" />
        ) : (
          <div className="flex gap-2">
            <Input
              id="filePath"
              value={state.filePath}
              aria-required
              aria-invalid={!!errors.filePath}
              className={fieldErrorClass(!!errors.filePath)}
              onChange={(e) => onChange({ filePath: e.target.value })}
            />
            {onBrowse && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={async () => {
                  const path = await onBrowse()
                  if (path) onChange({ filePath: path })
                }}
              >
                {t('connection.browse')}
              </Button>
            )}
          </div>
        )}
      </FormField>
    </div>
  )
}
