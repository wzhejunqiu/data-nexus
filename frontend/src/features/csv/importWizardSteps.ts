export type ImportTargetMode = 'new' | 'existing'

export type ImportStep = 'source' | 'target' | 'mapping' | 'progress' | 'result'

export const IMPORT_STEPS: ImportStep[] = ['source', 'target', 'mapping', 'progress', 'result']

/** User-facing wizard steps shown in the step indicator (excludes progress/result). */
export const IMPORT_WIZARD_STEPS: ImportStep[] = ['source', 'target', 'mapping']

export function getImportSteps(): ImportStep[] {
  return IMPORT_STEPS
}

export function getImportWizardSteps(): ImportStep[] {
  return IMPORT_WIZARD_STEPS
}

export function isInteractiveImportStep(step: ImportStep): boolean {
  return IMPORT_WIZARD_STEPS.includes(step)
}

export function stepLabelKey(step: ImportStep): string {
  switch (step) {
    case 'source':
      return 'csv.stepSource'
    case 'target':
      return 'csv.stepTarget'
    case 'mapping':
      return 'csv.stepMapping'
    case 'progress':
      return 'csv.stepImportProgress'
    case 'result':
      return 'csv.stepImportResult'
  }
}
