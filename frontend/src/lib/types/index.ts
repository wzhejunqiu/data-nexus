export type DriverType = 'sqlite' | 'postgres' | 'mysql'
export type SecretsBackend = 'keychain' | 'vault'

export interface AppError {
  code: string
  message: string
  details?: Record<string, unknown>
}

export interface SQLiteConfig {
  filePath: string
  readOnly: boolean
  wal?: boolean
}

export interface PostgresConfig {
  host: string
  port: number
  database: string
  user: string
  sslMode?: string
  schema?: string
  readOnly: boolean
}

export interface MySQLConfig {
  host: string
  port: number
  database: string
  user: string
  tls: boolean
  tlsSkipVerify?: boolean
  readOnly: boolean
}

export interface DriverConfig {
  type: DriverType
  sqlite?: SQLiteConfig
  postgres?: PostgresConfig
  mysql?: MySQLConfig
}

export interface ConnectRequest {
  filePath: string
  readOnly: boolean
  wal?: boolean
}

export interface RemoteConnectRequest {
  type: DriverType
  name?: string
  password: string
  postgres?: PostgresConfig
  mysql?: MySQLConfig
  open?: boolean
}

export interface TestConnectionRequest {
  type: DriverType
  password: string
  postgres?: PostgresConfig
  mysql?: MySQLConfig
}

export interface SQLiteSettingsUpdate {
  readOnly: boolean
  wal: boolean
}

export interface PostgresSettingsUpdate {
  host: string
  port: number
  database: string
  user: string
  password?: string | null
  sslMode: string
  schema: string
  readOnly: boolean
}

export interface MySQLSettingsUpdate {
  host: string
  port: number
  database: string
  user: string
  password?: string | null
  tls: boolean
  tlsSkipVerify: boolean
  readOnly: boolean
}

export interface SavedConnection {
  id: string
  name: string
  type: DriverType
  config: DriverConfig
  secretsBackend?: SecretsBackend
  createdAt: string
  updatedAt: string
  lastUsedAt: string
}

export interface Connection {
  id: string
  type: DriverType
  displayName: string
  config: DriverConfig
  connectedAt: string
}

export type ConnectionStatus = 'open' | 'closed'

export interface ConnectionListItem {
  id: string
  name: string
  type: DriverType
  config: DriverConfig
  secretsBackend?: SecretsBackend
  status: ConnectionStatus
  connectedAt?: string
  lastUsedAt: string
}

export interface ConnectionListView {
  items: ConnectionListItem[]
}

export type TableType = 'table' | 'view'

export interface TableInfo {
  name: string
  type: TableType
  schema?: string | null
  rowCount?: number | null
}

export interface TableList {
  items: TableInfo[]
}

export interface ColumnInfo {
  name: string
  dataType: string
  nativeType?: string
  nullable: boolean
  primaryKey: boolean
  defaultValue?: string | null
  position: number
}

export interface IndexInfo {
  name: string
  columns: string[]
  unique: boolean
  primary: boolean
}

export interface TableSchema {
  name: string
  type: TableType
  schema?: string | null
  columns: ColumnInfo[]
  indexes: IndexInfo[]
}

export interface ColumnMeta {
  name: string
  dataType: string
}

export interface PaginationMeta {
  page: number
  pageSize: number
  totalRows: number
  totalPages: number
}

export interface PaginatedTableData {
  columns: ColumnMeta[]
  rows: Record<string, unknown>[]
  pagination: PaginationMeta
}

export type FilterOperator =
  | 'eq'
  | 'ne'
  | 'gt'
  | 'gte'
  | 'lt'
  | 'lte'
  | 'like'
  | 'is_null'
  | 'is_not_null'
  | 'in'

export interface RowFilter {
  column: string
  operator: FilterOperator
  value?: string | null
  values?: string[]
}

export interface BrowseRowsRequest {
  connectionId: string
  tableName: string
  page: number
  pageSize: number
  sort: string
  order: 'asc' | 'desc'
  filters?: RowFilter[]
  search?: string
}

export interface ValueCount {
  value: string
  count: number
}

export interface ColumnProfile {
  name: string
  distinctCount?: number | null
  nullPercent?: number | null
  minValue?: string | null
  maxValue?: string | null
  topValues?: ValueCount[]
  isLowCardinality: boolean
}

export interface TableProfile {
  tableName: string
  sampledRows: number
  totalRows?: number | null
  isSampled: boolean
  columns: ColumnProfile[]
}

export interface FTSInfo {
  enabled: boolean
  schema?: string
  ftsTableName?: string
  contentTable?: string
}

export interface AttachedDatabase {
  alias: string
  filePath: string
}

export interface CannedQuery {
  id: string
  name: string
  sql: string
  connectionId?: string | null
  tags?: string[]
  createdAt: string
  updatedAt: string
}

export interface CannedQueryList {
  items: CannedQuery[]
}

export interface SaveCannedQueryRequest {
  id?: string
  name: string
  sql: string
  connectionId?: string | null
  tags?: string[]
}

export interface ExecuteQueryRequest {
  connectionId: string
  sql: string
  params?: unknown[]
  maxRows?: number
}

export interface QueryResponse {
  kind: 'result' | 'exec'
  columns?: ColumnMeta[]
  rows?: Record<string, unknown>[]
  rowCount?: number
  truncated?: boolean
  rowsAffected?: number
  lastInsertId?: number
  durationMs: number
}

export type StatementKind = 'query' | 'write'

export type SqlExecutionKind = 'result' | 'exec'

export interface SqlExecutionRecord {
  id?: number
  connectionId: string
  sql: string
  kind: SqlExecutionKind
  effectRows: number
  durationMs: number
  executedAt: string
}

export interface SqlExecutionList {
  items: SqlExecutionRecord[]
}

export interface VersionInfo {
  version: string
  platform: string
  arch: string
}

export * from './csv'
export * from './edit'
