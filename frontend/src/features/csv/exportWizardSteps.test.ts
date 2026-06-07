import { describe, expect, it } from 'vitest'
import { getExportSteps, isTableExport } from './exportWizardSteps'
import { useExportStore } from '@/stores/exportStore'

describe('getExportSteps', () => {
  it('includes scope for table exports', () => {
    expect(getExportSteps('table-page')[0]).toBe('scope')
    expect(getExportSteps('table-all')).toContain('result')
    expect(getExportSteps('table-all').at(-1)).toBe('result')
  })

  it('skips scope for query result export', () => {
    const steps = getExportSteps('query-result')
    expect(steps[0]).toBe('columns')
    expect(steps).not.toContain('scope')
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
