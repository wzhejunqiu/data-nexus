import {
  DetectFTSTable,
  GetTableProfile,
  GetTableSchema,
  ListNamespaces,
  ListSchemas,
  ListTables,
} from '../../../wailsjs/go/wails/SchemaService'
import type {
  FTSInfo,
  ListTablesOptions,
  NamespaceList,
  SchemaList,
  TableList,
  TableProfile,
  TableSchema,
} from '../types'
import { mapWailsError } from './errors'

async function wrap<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    throw mapWailsError(err)
  }
}

export const schemaApi = {
  listNamespaces: (connectionId: string) =>
    wrap(() => ListNamespaces(connectionId) as Promise<NamespaceList>),
  listSchemas: (connectionId: string, database: string) =>
    wrap(() => ListSchemas(connectionId, database) as Promise<SchemaList>),
  listTables: (connectionId: string, opts: ListTablesOptions = {}) =>
    wrap(() => ListTables(connectionId, opts) as Promise<TableList>),
  getTableSchema: (connectionId: string, tableName: string) =>
    wrap(() => GetTableSchema(connectionId, tableName) as Promise<TableSchema>),
  getTableProfile: (connectionId: string, tableName: string) =>
    wrap(() => GetTableProfile(connectionId, tableName) as Promise<TableProfile>),
  detectFTS: (connectionId: string, tableName: string) =>
    wrap(() => DetectFTSTable(connectionId, tableName) as Promise<FTSInfo>),
}
