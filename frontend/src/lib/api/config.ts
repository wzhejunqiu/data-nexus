import { GetConfig, GetConfigPath, UpdateConfig } from '../../../wailsjs/go/wails/ConfigService'
import { config } from '../../../wailsjs/go/models'
import { mapWailsError } from './errors'

export interface AppConfig {
  log: {
    level: string
    output: string
    file: {
      path: string
      max_size_mb: number
      max_backups: number
      max_age_days: number
      compress: boolean
    }
  }
}

function fromWailsConfig(c: config.Config): AppConfig {
  return {
    log: {
      level: c.Log.Level,
      output: c.Log.Output,
      file: {
        path: c.Log.File.Path,
        max_size_mb: c.Log.File.MaxSizeMB,
        max_backups: c.Log.File.MaxBackups,
        max_age_days: c.Log.File.MaxAgeDays,
        compress: c.Log.File.Compress,
      },
    },
  }
}

function toWailsConfig(cfg: AppConfig): config.Config {
  return config.Config.createFrom({
    Log: {
      Level: cfg.log.level,
      Output: cfg.log.output,
      File: {
        Path: cfg.log.file.path,
        MaxSizeMB: cfg.log.file.max_size_mb,
        MaxBackups: cfg.log.file.max_backups,
        MaxAgeDays: cfg.log.file.max_age_days,
        Compress: cfg.log.file.compress,
      },
    },
  })
}

export const configApi = {
  getConfig: async (): Promise<AppConfig> => {
    try {
      return fromWailsConfig(await GetConfig())
    } catch (err) {
      throw mapWailsError(err)
    }
  },
  updateConfig: async (cfg: AppConfig) => {
    try {
      await UpdateConfig(toWailsConfig(cfg))
    } catch (err) {
      throw mapWailsError(err)
    }
  },
  getConfigPath: async (): Promise<string> => {
    try {
      return (await GetConfigPath()) as string
    } catch (err) {
      throw mapWailsError(err)
    }
  },
}
