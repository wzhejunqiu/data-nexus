import { useCallback, useImperativeHandle, useRef, useState, forwardRef } from 'react'
import { useTranslation } from 'react-i18next'
import type { ConnectionListItem, DriverType } from '@/lib/types'
import { DialectSidebar } from './DialectSidebar'
import { GeneralSection } from './sections/GeneralSection'
import { SecuritySection } from './sections/SecuritySection'
import { AdvancedSection } from './sections/AdvancedSection'
import {
  connectionFormStateFromItem,
  defaultConnectionFormState,
  type ConnectionFormState,
} from './connectionFormDefaults'
import {
  connectionFailureFieldErrors,
  firstErrorField,
  sectionForField,
  validateConnectionField,
  validateConnectionForm,
  type ConnectionFormErrors,
  type ConnectionFormField,
} from './connectionFormValidation'

type SectionKey = 'general' | 'security' | 'advanced'

export type ConnectionFormHandle = {
  submit: () => void
  test: () => void
  highlightConnectionFailure: (reason: string) => void
}

function CollapsibleSection({
  title,
  open,
  onToggle,
  children,
  sectionRef,
}: {
  title: string
  open: boolean
  onToggle: () => void
  children: React.ReactNode
  sectionRef?: React.RefObject<HTMLDivElement>
}) {
  return (
    <div
      ref={sectionRef as React.RefObject<HTMLDivElement>}
      className="border-b border-border last:border-b-0"
    >
      <button
        type="button"
        onClick={onToggle}
        className="flex w-full items-center gap-2 py-2 text-left text-sm font-medium"
      >
        <span className="text-muted">{open ? '▼' : '▶'}</span>
        {title}
      </button>
      {open && <div className="pb-3">{children}</div>}
    </div>
  )
}

export const ConnectionForm = forwardRef<
  ConnectionFormHandle,
  {
    mode: 'create' | 'edit'
    initialItem?: ConnectionListItem
    initialDialect?: DriverType
    readOnlyDefault?: boolean
    walDefault?: boolean
    dialectLocked?: boolean
    onBrowse?: () => Promise<string | undefined>
    onSubmit: (state: ConnectionFormState) => void
    onTest?: (state: ConnectionFormState) => void
    showTest?: boolean
    testPending?: boolean
  }
>(function ConnectionForm(
  {
    mode,
    initialItem,
    initialDialect,
    readOnlyDefault,
    walDefault,
    dialectLocked,
    onBrowse,
    onSubmit,
    onTest,
    showTest,
    testPending,
  },
  ref,
) {
  const { t } = useTranslation()
  const [state, setState] = useState<ConnectionFormState>(() => {
    if (initialItem) return connectionFormStateFromItem(initialItem)
    const base = defaultConnectionFormState(initialDialect ?? 'sqlite')
    if (readOnlyDefault !== undefined) base.readOnly = readOnlyDefault
    if (walDefault !== undefined) base.wal = walDefault
    return base
  })
  const [errors, setErrors] = useState<ConnectionFormErrors>({})
  const [openSections, setOpenSections] = useState<Record<SectionKey, boolean>>({
    general: true,
    security: false,
    advanced: false,
  })
  const generalRef = useRef<HTMLDivElement>(null)
  const securityRef = useRef<HTMLDivElement>(null)
  const advancedRef = useRef<HTMLDivElement>(null)

  const dialect = dialectLocked && initialItem ? initialItem.type : state.dialect

  const focusFirstError = useCallback(
    (next: ConnectionFormErrors) => {
      const first = firstErrorField(next)
      if (!first) return
      const section = sectionForField(first, dialect)
      setOpenSections((s) => ({ ...s, [section]: true }))
      const refMap = { general: generalRef, security: securityRef, advanced: advancedRef }
      window.requestAnimationFrame(() => {
        refMap[section].current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
        document.getElementById(first)?.focus()
      })
    },
    [dialect],
  )

  const patch = useCallback((p: Partial<ConnectionFormState>) => {
    setState((prev) => ({ ...prev, ...p }))
    setErrors((prev) => {
      const next = { ...prev }
      if ('displayName' in p) delete next.displayName
      if ('filePath' in p) delete next.filePath
      if ('password' in p) delete next.password
      if ('postgres' in p) {
        delete next.host
        delete next.port
        delete next.database
        delete next.user
      }
      if ('mysql' in p) {
        delete next.host
        delete next.port
        delete next.database
        delete next.user
      }
      return next
    })
  }, [])

  const handleFieldBlur = useCallback(
    (field: ConnectionFormField) => {
      const code = validateConnectionField(field, state, mode, dialect)
      setErrors((prev) => {
        const next = { ...prev }
        if (code) {
          next[field] = code
        } else {
          delete next[field]
        }
        return next
      })
    },
    [state, mode, dialect],
  )

  const highlightConnectionFailure = useCallback(
    (reason: string) => {
      const failureErrors = connectionFailureFieldErrors(reason)
      if (Object.keys(failureErrors).length === 0) return
      setErrors((prev) => ({ ...prev, ...failureErrors }))
      setOpenSections((s) => ({ ...s, general: true }))
      focusFirstError(failureErrors)
    },
    [focusFirstError],
  )

  const runValidation = useCallback(() => {
    const next = validateConnectionForm(state, mode, dialect)
    setErrors(next)
    focusFirstError(next)
    return Object.keys(next).length === 0
  }, [state, mode, dialect, focusFirstError])

  const submit = useCallback(() => {
    if (!runValidation()) return
    onSubmit({ ...state, dialect })
  }, [runValidation, onSubmit, state, dialect])

  const test = useCallback(() => {
    if (!runValidation()) return
    onTest?.({ ...state, dialect })
  }, [runValidation, onTest, state, dialect])

  useImperativeHandle(ref, () => ({ submit, test, highlightConnectionFailure }), [
    submit,
    test,
    highlightConnectionFailure,
  ])

  const showSecurity = dialect === 'postgres' || dialect === 'mysql'

  return (
    <div className="flex max-h-[60vh] gap-4">
      <DialectSidebar
        dialect={dialect}
        disabled={dialectLocked || mode === 'edit'}
        onDialectChange={(d) => patch({ dialect: d })}
      />
      <div className="min-w-0 flex-1 overflow-y-auto pr-1">
        <CollapsibleSection
          title={t('connectionForm.section.general')}
          open={openSections.general}
          onToggle={() => setOpenSections((s) => ({ ...s, general: !s.general }))}
          sectionRef={generalRef}
        >
          <GeneralSection
            dialect={dialect}
            mode={mode}
            state={state}
            errors={errors}
            onChange={patch}
            onBrowse={onBrowse}
            onFieldBlur={handleFieldBlur}
          />
        </CollapsibleSection>
        {showSecurity && (
          <CollapsibleSection
            title={t('connectionForm.section.security')}
            open={openSections.security}
            onToggle={() => setOpenSections((s) => ({ ...s, security: !s.security }))}
            sectionRef={securityRef}
          >
            <SecuritySection dialect={dialect} state={state} onChange={patch} />
          </CollapsibleSection>
        )}
        <CollapsibleSection
          title={t('connectionForm.section.advanced')}
          open={openSections.advanced}
          onToggle={() => setOpenSections((s) => ({ ...s, advanced: !s.advanced }))}
          sectionRef={advancedRef}
        >
          <AdvancedSection dialect={dialect} state={state} onChange={patch} />
        </CollapsibleSection>
        {showTest && onTest && dialect !== 'sqlite' && (
          <div className="mt-4">
            <button
              type="button"
              className="text-sm text-accent hover:underline disabled:opacity-50"
              disabled={testPending}
              onClick={test}
            >
              {t('connectionForm.testConnection')}
            </button>
          </div>
        )}
      </div>
    </div>
  )
})

export type { ConnectionFormState }
