import type { ConnectionListItem, DriverType, MySQLConfig, PostgresConfig } from '@/lib/types'

export const POSTGRES_SSL_MODES = [
  'disable',
  'allow',
  'prefer',
  'require',
  'verify-ca',
  'verify-full',
] as const

export const POSTGRES_CLIENT_ENCODINGS = ['UTF8', 'LATIN1', 'SQL_ASCII', 'GBK'] as const

export const MYSQL_CHARSETS = ['utf8mb4', 'utf8', 'latin1', 'gbk'] as const

export const MYSQL_COLLATIONS: Record<string, string[]> = {
  utf8mb4: ['utf8mb4_unicode_ci', 'utf8mb4_general_ci', 'utf8mb4_0900_ai_ci'],
  utf8: ['utf8_general_ci', 'utf8_unicode_ci'],
  latin1: ['latin1_swedish_ci'],
  gbk: ['gbk_chinese_ci'],
}

export const MYSQL_STORAGE_ENGINES = ['', 'InnoDB', 'MyISAM'] as const

export interface ConnectionFormState {
  dialect: DriverType
  displayName: string
  password: string
  filePath: string
  readOnly: boolean
  wal: boolean
  postgres: PostgresConfig
  mysql: MySQLConfig
}

export function defaultPostgres(): PostgresConfig {
  return {
    host: 'localhost',
    port: 5432,
    database: '',
    user: '',
    sslMode: 'disable',
    schema: 'public',
    clientEncoding: 'UTF8',
    readOnly: false,
  }
}

export function defaultMySQL(): MySQLConfig {
  return {
    host: 'localhost',
    port: 3306,
    database: '',
    user: '',
    tls: false,
    tlsSkipVerify: false,
    charset: 'utf8mb4',
    collation: 'utf8mb4_unicode_ci',
    defaultStorageEngine: 'InnoDB',
    readOnly: false,
  }
}

export function defaultConnectionFormState(dialect: DriverType = 'sqlite'): ConnectionFormState {
  return {
    dialect,
    displayName: '',
    password: '',
    filePath: '',
    readOnly: false,
    wal: false,
    postgres: defaultPostgres(),
    mysql: defaultMySQL(),
  }
}

export function connectionFormStateFromItem(item: ConnectionListItem): ConnectionFormState {
  return {
    dialect: item.type,
    displayName: item.name,
    password: '',
    filePath: item.config.sqlite?.filePath ?? '',
    readOnly:
      item.config.sqlite?.readOnly ??
      item.config.postgres?.readOnly ??
      item.config.mysql?.readOnly ??
      false,
    wal: item.config.sqlite?.wal ?? false,
    postgres: {
      ...defaultPostgres(),
      ...item.config.postgres,
      schema: item.config.postgres?.schema ?? 'public',
      sslMode: item.config.postgres?.sslMode ?? 'disable',
      clientEncoding: item.config.postgres?.clientEncoding ?? 'UTF8',
    },
    mysql: {
      ...defaultMySQL(),
      ...item.config.mysql,
      charset: item.config.mysql?.charset ?? 'utf8mb4',
      collation: item.config.mysql?.collation ?? 'utf8mb4_unicode_ci',
      defaultStorageEngine: item.config.mysql?.defaultStorageEngine ?? 'InnoDB',
    },
  }
}
