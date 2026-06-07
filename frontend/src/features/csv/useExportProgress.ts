import { useEffect, useState } from 'react'
import { EventsOn } from '../../../wailsjs/runtime/runtime'

export interface ExportProgressState {
  exported: number
}

export function useExportProgress(exportId: string | null): ExportProgressState | null {
  const [progress, setProgress] = useState<ExportProgressState | null>(null)

  useEffect(() => {
    if (!exportId) {
      setProgress(null)
      return
    }
    setProgress(null)
    return EventsOn('export:progress', (payload: { exportId?: string; exported?: number }) => {
      if (payload?.exportId !== exportId) return
      setProgress({ exported: payload.exported ?? 0 })
    })
  }, [exportId])

  return progress
}
