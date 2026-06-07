import { describe, expect, it } from 'vitest'
import {
  getExportSteps,
  getExportWizardSteps,
  isInteractiveExportStep,
  isTableExport,
  stepLabelKey,
} from './exportWizardSteps'
import { useExportStore } from '@/stores/exportStore'

describe('getExportSteps', () => {
  it('returns full export flow including progress and result', () => {
    const steps = getExportSteps()
    expect(steps).toEqual(['options', 'destination', 'progress', 'result'])
    expect(steps).toHaveLength(4)
  })
})

describe('getExportWizardSteps', () => {
  it('returns only interactive steps for the step indicator', () => {
    expect(getExportWizardSteps()).toEqual(['options', 'destination'])
    expect(isInteractiveExportStep('options')).toBe(true)
    expect(isInteractiveExportStep('destination')).toBe(true)
    expect(isInteractiveExportStep('progress')).toBe(false)
    expect(isInteractiveExportStep('result')).toBe(false)
  })
})

describe('stepLabelKey', () => {
  it('maps export steps to i18n keys', () => {
    expect(stepLabelKey('options')).toBe('csv.stepExportOptions')
    expect(stepLabelKey('destination')).toBe('csv.stepDestination')
    expect(stepLabelKey('result')).toBe('csv.stepResult')
  })
})

describe('isTableExport', () => {
  it('identifies table sources', () => {
    expect(isTableExport('table-page')).toBe(true)
    expect(isTableExport('table-all')).toBe(true)
    expect(isTableExport('query-result')).toBe(false)
  })
})

describe('useExportStore', () => {
  it('opens and closes export session', () => {
    useExportStore.getState().closeExport()
    expect(useExportStore.getState().session).toBeNull()

    useExportStore.getState().openExport({
      source: 'query-result',
      connectionId: 'c1',
      availableColumns: ['a'],
      rows: [{ a: 1 }],
    })
    expect(useExportStore.getState().session?.source).toBe('query-result')

    useExportStore.getState().closeExport()
    expect(useExportStore.getState().session).toBeNull()
  })
})
