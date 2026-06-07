import {
  GetVersion,
  SetActiveConnection,
  SetWindowTitle,
} from '../../../wailsjs/go/wails/AppService'
import type { VersionInfo } from '../types'
import { mapWailsError } from './errors'

async function wrap<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    throw mapWailsError(err)
  }
}

export const appApi = {
  getVersion: () => wrap(() => GetVersion() as Promise<VersionInfo>),
  setWindowTitle: (title: string) => wrap(() => SetWindowTitle(title)),
  setActiveConnection: (connectionId: string) => wrap(() => SetActiveConnection(connectionId)),
}
