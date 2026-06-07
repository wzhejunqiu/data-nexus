export interface AppError {
  code: string
  message: string
  details?: Record<string, unknown>
}

export interface SQLiteConfig {
  filePath: string
  readOnly: boolean
}

export interface DriverConfig {
  type: 'sqlite'
  sqlite?: SQLiteConfig
}

export interface ConnectRequest {
  filePath: string
  readOnly: boolean
}

export interface SavedConnection {
  id: string
  name: string
  type: 'sqlite'
  config: DriverConfig
  createdAt: string
  updatedAt: string
  lastUsedAt: string
}

export interface Connection {
  id: string
  type: 'sqlite'
  displayName: string
  config: DriverConfig
  connectedAt: string
}

export type ConnectionStatus = 'open' | 'closed'

export interface ConnectionListItem {
  id: string
  name: string
  type: 'sqlite'
  config: DriverConfig
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

export interface BrowseRowsRequest {
  connectionId: string
  tableName: string
  page: number
  pageSize: number
  sort: string
  order: 'asc' | 'desc'
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

export interface VersionInfo {
  version: string
  platform: string
  arch: string
}
