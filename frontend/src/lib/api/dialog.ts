import { OpenDatabaseFile, OpenCSVFile, SaveFile } from '../../../wailsjs/go/wails/DialogService'
import type { FileFilter } from '../types'
import { mapWailsError } from './errors'

export const dialogApi = {
  openDatabaseFile: async () => {
    try {
      return (await OpenDatabaseFile()) as string
    } catch (err) {
      throw mapWailsError(err)
    }
  },
  openCSVFile: async () => {
    try {
      return (await OpenCSVFile()) as string
    } catch (err) {
      throw mapWailsError(err)
    }
  },
  saveFile: async (defaultName: string, filters: FileFilter[]) => {
    try {
      return (await SaveFile(defaultName, filters)) as string
    } catch (err) {
      throw mapWailsError(err)
    }
  },
}
