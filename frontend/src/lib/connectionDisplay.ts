import type { ConnectionListItem } from '@/lib/types'

export function connectionTypeLabel(type: ConnectionListItem['type']): string {
  switch (type) {
    case 'postgres':
      return 'PG'
    case 'mysql':
      return 'MY'
    default:
      return 'SQL'
  }
}

export function connectionSubtitle(item: ConnectionListItem): string {
  if (item.config.sqlite?.filePath) {
    return item.config.sqlite.filePath
  }
  if (item.config.postgres) {
    const p = item.config.postgres
    const port = p.port || 5432
    return `${p.user}@${p.host}:${port}/${p.database}`
  }
  if (item.config.mysql) {
    const m = item.config.mysql
    const port = m.port || 3306
    const base = `${m.user}@${m.host}:${port}`
    const db = m.database?.trim()
    return db ? `${base}/${db}` : base
  }
  return ''
}

export function isRemoteConnection(item: Pick<ConnectionListItem, 'type'>): boolean {
  return item.type === 'postgres' || item.type === 'mysql'
}
