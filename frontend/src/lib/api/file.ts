import { WriteTextFile } from '../../../wailsjs/go/wails/FileService'
import { mapWailsError } from './errors'

export const fileApi = {
  writeTextFile: async (path: string, content: string, encoding: string) => {
    try {
      await WriteTextFile(path, content, encoding)
    } catch (err) {
      throw mapWailsError(err)
    }
  },
}
