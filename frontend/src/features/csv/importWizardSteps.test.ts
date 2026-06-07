import { describe, expect, it } from 'vitest'
import {
  getImportSteps,
  getImportWizardSteps,
  isInteractiveImportStep,
  stepLabelKey,
} from './importWizardSteps'
import { useImportStore } from '@/stores/importStore'

describe('getImportSteps', () => {
  it('returns full import flow including progress and result', () => {
    const steps = getImportSteps()
    expect(steps).toEqual(['source', 'target', 'mapping', 'progress', 'result'])
    expect(steps).toHaveLength(5)
  })
})

describe('getImportWizardSteps', () => {
  it('returns only interactive steps for the step indicator', () => {
    expect(getImportWizardSteps()).toEqual(['source', 'target', 'mapping'])
    expect(isInteractiveImportStep('source')).toBe(true)
    expect(isInteractiveImportStep('mapping')).toBe(true)
    expect(isInteractiveImportStep('progress')).toBe(false)
    expect(isInteractiveImportStep('result')).toBe(false)
  })
})

describe('stepLabelKey', () => {
  it('maps import steps to i18n keys', () => {
    expect(stepLabelKey('source')).toBe('csv.stepSource')
    expect(stepLabelKey('mapping')).toBe('csv.stepMapping')
    expect(stepLabelKey('result')).toBe('csv.stepImportResult')
  })
})

describe('useImportStore', () => {
  it('opens and closes import session', () => {
    useImportStore.getState().closeImport()
    expect(useImportStore.getState().session).toBeNull()

    useImportStore.getState().openImport({ connectionId: 'c1', defaultTable: 'items' })
    expect(useImportStore.getState().session?.connectionId).toBe('c1')
    expect(useImportStore.getState().session?.defaultTable).toBe('items')

    useImportStore.getState().closeImport()
    expect(useImportStore.getState().session).toBeNull()
  })
})
