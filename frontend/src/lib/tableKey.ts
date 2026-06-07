import type { TableInfo } from './types'

/** Qualified table name for API calls (attached DBs use alias.table). */
export function tableKey(item: Pick<TableInfo, 'name' | 'schema'>): string {
  if (item.schema && item.schema !== 'main') {
    return `${item.schema}.${item.name}`
  }
  return item.name
}
