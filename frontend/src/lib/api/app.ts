import { GetVersion } from '../../../wailsjs/go/wails/AppService'
import type { VersionInfo } from '../types'
import { mapWailsError } from './errors'

export const appApi = {
  getVersion: async () => {
    try {
      return (await GetVersion()) as VersionInfo
    } catch (err) {
      throw mapWailsError(err)
    }
  },
}
