import type { ExportSource } from '@/stores/exportStore'

export type ExportStep = 'options' | 'destination' | 'progress' | 'result'

export const EXPORT_STEPS: ExportStep[] = ['options', 'destination', 'progress', 'result']

/** User-facing wizard steps shown in the step indicator (excludes progress/result). */
export const EXPORT_WIZARD_STEPS: ExportStep[] = ['options', 'destination']

export function getExportSteps(): ExportStep[] {
  return EXPORT_STEPS
}

export function getExportWizardSteps(): ExportStep[] {
  return EXPORT_WIZARD_STEPS
}

export function isInteractiveExportStep(step: ExportStep): boolean {
  return EXPORT_WIZARD_STEPS.includes(step)
}

export function stepLabelKey(step: ExportStep): string {
  switch (step) {
    case 'options':
      return 'csv.stepExportOptions'
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
