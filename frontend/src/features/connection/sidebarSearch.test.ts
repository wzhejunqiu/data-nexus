import { describe, expect, it } from 'vitest'
import {
  groupMatchesSearch,
  matchesConnectionName,
  matchesGroupName,
  normalizeSearchQuery,
} from './sidebarSearch'
import type { ConnectionGroupNode, ConnectionListItem } from '@/lib/types'

function conn(id: string, name: string): ConnectionListItem {
  return {
    id,
    name,
    type: 'sqlite',
    config: { type: 'sqlite', sqlite: { filePath: '/tmp/a.db', readOnly: false } },
    status: 'closed',
    lastUsedAt: '2026-01-01T00:00:00Z',
  }
}

describe('sidebarSearch', () => {
  it('normalizes query', () => {
    expect(normalizeSearchQuery('  Foo ')).toBe('foo')
  })

  it('matches connection name', () => {
    expect(matchesConnectionName(conn('1', 'Production DB'), 'prod')).toBe(true)
    expect(matchesConnectionName(conn('1', 'Production DB'), 'staging')).toBe(false)
  })

  it('matches group name', () => {
    expect(matchesGroupName('My Work', 'work')).toBe(true)
    expect(matchesGroupName('My Work', 'archive')).toBe(false)
  })

  it('matches group when child connection matches', () => {
    const node: ConnectionGroupNode = {
      id: 'g1',
      name: 'Archive',
      connections: [conn('c1', 'staging.db')],
      childGroups: [],
    }
    expect(groupMatchesSearch(node, 'staging')).toBe(true)
    expect(groupMatchesSearch(node, 'nomatch')).toBe(false)
  })

  it('matches nested group child', () => {
    const node: ConnectionGroupNode = {
      id: 'g1',
      name: 'Root',
      connections: [],
      childGroups: [
        {
          id: 'g2',
          name: 'Nested',
          connections: [conn('c1', 'mysql-prod')],
          childGroups: [],
        },
      ],
    }
    expect(groupMatchesSearch(node, 'mysql')).toBe(true)
  })
})
