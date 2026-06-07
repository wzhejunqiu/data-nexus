import { useEffect, useState } from 'react'
import { EventsOn } from '../../../wailsjs/runtime/runtime'

export interface ExportProgressState {
  exported: number
}

export function useExportProgress(exportId: string | null): ExportProgressState | null {
  const [progress, setProgress] = useState<{ exportId: string; exported: number } | null>(null)

  useEffect(() => {
    if (!exportId) return
    return EventsOn('export:progress', (payload: { exportId?: string; exported?: number }) => {
      if (payload?.exportId !== exportId) return
      setProgress({ exportId, exported: payload.exported ?? 0 })
    })
  }, [exportId])

  if (!exportId || progress?.exportId !== exportId) return null
  return { exported: progress.exported }
}
