import { describe, expect, it } from 'vitest'
import { connectionSubtitle } from './connectionDisplay'
import type { ConnectionListItem } from './types'

describe('connectionSubtitle', () => {
  it('omits mysql database segment when empty', () => {
    const item: ConnectionListItem = {
      id: 'my-1',
      name: 'local-mysql',
      type: 'mysql',
      config: {
        type: 'mysql',
        mysql: {
          host: 'localhost',
          port: 3306,
          database: '',
          user: 'root',
          tls: false,
          readOnly: false,
        },
      },
      status: 'closed',
      lastUsedAt: '2026-01-01T00:00:00Z',
    }
    expect(connectionSubtitle(item)).toBe('root@localhost:3306')
  })

  it('includes mysql database when set', () => {
    const item: ConnectionListItem = {
      id: 'my-1',
      name: 'local-mysql',
      type: 'mysql',
      config: {
        type: 'mysql',
        mysql: {
          host: 'localhost',
          port: 3306,
          database: 'app',
          user: 'root',
          tls: false,
          readOnly: false,
        },
      },
      status: 'closed',
      lastUsedAt: '2026-01-01T00:00:00Z',
    }
    expect(connectionSubtitle(item)).toBe('root@localhost:3306/app')
  })
})
