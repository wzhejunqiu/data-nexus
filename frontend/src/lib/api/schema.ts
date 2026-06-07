import { GetTableSchema, ListTables } from '../../../wailsjs/go/wails/SchemaService'
import type { TableList, TableSchema } from '../types'
import { mapWailsError } from './errors'

async function wrap<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    throw mapWailsError(err)
  }
}

export const schemaApi = {
  listTables: (connectionId: string) => wrap(() => ListTables(connectionId) as Promise<TableList>),
  getTableSchema: (connectionId: string, tableName: string) =>
    wrap(() => GetTableSchema(connectionId, tableName) as Promise<TableSchema>),
}
