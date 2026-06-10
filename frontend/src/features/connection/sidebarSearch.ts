import type { ConnectionGroupNode, ConnectionListItem, DriverType } from '@/lib/types'

export function normalizeSearchQuery(query: string): string {
  return query.trim().toLowerCase()
}

export function matchesConnectionName(item: ConnectionListItem, query: string): boolean {
  const q = normalizeSearchQuery(query)
  if (!q) return true
  return item.name.toLowerCase().includes(q)
}

export function matchesGroupName(name: string, query: string): boolean {
  const q = normalizeSearchQuery(query)
  if (!q) return true
  return name.toLowerCase().includes(q)
}

/** Namespace/schema name match for schema tree filtering (table match handled in component). */
export function matchesSchemaSearch(
  name: string,
  _driverType: DriverType,
  query: string,
  _connectionId: string,
  _ctx?: { database?: string; schema?: string },
): boolean {
  const q = normalizeSearchQuery(query)
  if (!q) return true
  return name.toLowerCase().includes(q)
}

export function groupMatchesSearch(node: ConnectionGroupNode, query: string): boolean {
  const q = normalizeSearchQuery(query)
  if (!q) return true
  if (matchesGroupName(node.name, q)) return true
  for (const conn of node.connections ?? []) {
    if (matchesConnectionName(conn, q)) return true
  }
  for (const child of node.childGroups ?? []) {
    if (groupMatchesSearch(child, q)) return true
  }
  return false
}
