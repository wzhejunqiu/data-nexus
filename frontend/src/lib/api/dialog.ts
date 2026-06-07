import { OpenDatabaseFile } from '../../../wailsjs/go/wails/DialogService'
import { mapWailsError } from './errors'

export const dialogApi = {
  openDatabaseFile: async () => {
    try {
      return (await OpenDatabaseFile()) as string
    } catch (err) {
      throw mapWailsError(err)
    }
  },
}
