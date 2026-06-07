export type SqlDialect = 'sqlite' | 'mysql' | 'postgresql' | 'sql'

export function connectionTypeToSqlDialect(type?: string): SqlDialect {
  switch (type) {
    case 'mysql':
      return 'mysql'
    case 'postgresql':
    case 'postgres':
      return 'postgresql'
    case 'sqlite':
      return 'sqlite'
    default:
      return 'sql'
  }
}
