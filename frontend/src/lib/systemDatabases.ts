import type { DriverType } from '@/lib/types'

const MYSQL_SYSTEM = new Set(['information_schema', 'mysql', 'performance_schema', 'sys'])
const POSTGRES_SYSTEM = new Set(['template0', 'template1', 'postgres'])

export function isSystemDatabase(name: string, driverType: DriverType): boolean {
  const lower = name.toLowerCase()
  if (driverType === 'mysql') return MYSQL_SYSTEM.has(lower)
  if (driverType === 'postgres') return POSTGRES_SYSTEM.has(lower)
  return false
}
