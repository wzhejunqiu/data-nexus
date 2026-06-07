import type { ExportSource } from '@/stores/exportStore'

export type ExportStep = 'scope' | 'columns' | 'format' | 'destination' | 'progress' | 'result'

export function getExportSteps(source: ExportSource): ExportStep[] {
  if (source === 'query-result') {
    return ['columns', 'format', 'destination', 'progress', 'result']
  }
  return ['scope', 'columns', 'format', 'destination', 'progress', 'result']
}

export function stepLabelKey(step: ExportStep): string {
  switch (step) {
    case 'scope':
      return 'csv.stepScope'
    case 'columns':
      return 'csv.stepColumns'
    case 'format':
      return 'csv.stepFormat'
    case 'destination':
      return 'csv.stepDestination'
    case 'progress':
      return 'csv.stepProgress'
    case 'result':
      return 'csv.stepResult'
  }
}

export function isTableExport(source: ExportSource): boolean {
  return source === 'table-page' || source === 'table-all'
}
