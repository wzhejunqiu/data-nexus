import { describe, expect, it } from 'vitest'
import type { ConnectionGroupNode, ConnectionSidebarTree } from '@/lib/types'
import {
  collectGroupMemberMaps,
  defaultMemberItems,
  defaultRootItems,
  reorderMemberItems,
  reorderRootItems,
} from './sidebarOrder'

describe('sidebarOrder', () => {
  const tree: ConnectionSidebarTree = {
    groups: [
      {
        id: 'g1',
        name: 'G1',
        childGroups: [],
        connections: [
          {
            id: 'c1',
            name: 'a',
            type: 'sqlite',
            config: { type: 'sqlite', sqlite: { filePath: '/a', readOnly: false } },
            status: 'closed',
            lastUsedAt: '',
          },
        ],
        memberItems: [
          { memberType: 'connection', memberId: 'c1' },
          { memberType: 'group', memberId: 'g2' },
        ],
      },
      { id: 'g2', name: 'G2', childGroups: [], connections: [] },
    ],
    freeConnections: [
      {
        id: 'fc1',
        name: 'free',
        type: 'sqlite',
        config: { type: 'sqlite', sqlite: { filePath: '/f', readOnly: false } },
        status: 'closed',
        lastUsedAt: '',
      },
    ],
    rootItems: [
      { itemType: 'group', itemId: 'g1' },
      { itemType: 'connection', itemId: 'fc1' },
    ],
  }

  it('defaultRootItems prefers rootItems when present', () => {
    const items = defaultRootItems(tree)
    expect(items).toEqual(tree.rootItems)
  })

  it('defaultRootItems falls back to groups then free connections', () => {
    const items = defaultRootItems({ groups: tree.groups, freeConnections: tree.freeConnections })
    expect(items[0]).toEqual({ itemType: 'group', itemId: 'g1' })
    expect(items[1]).toEqual({ itemType: 'group', itemId: 'g2' })
    expect(items[2]).toEqual({ itemType: 'connection', itemId: 'fc1' })
  })

  it('defaultMemberItems prefers memberItems when present', () => {
    const node = tree.groups[0]
    expect(defaultMemberItems(node)).toEqual(node.memberItems)
  })

  it('defaultMemberItems falls back to connections and child groups', () => {
    const node: ConnectionGroupNode = {
      id: 'x',
      name: 'X',
      childGroups: [{ id: 'cg', name: 'CG', childGroups: [], connections: [] }],
      connections: tree.groups[0].connections,
    }
    const items = defaultMemberItems(node)
    expect(items[0].memberType).toBe('connection')
    expect(items[1].memberType).toBe('group')
  })

  it('collectGroupMemberMaps walks nested groups', () => {
    const maps = collectGroupMemberMaps(tree.groups)
    expect(maps.g1).toHaveLength(2)
    expect(maps.g2).toBeDefined()
  })

  it('reorderRootItems swaps siblings', () => {
    const items = defaultRootItems(tree)
    const next = reorderRootItems(items, 'fc1', 'g1')
    expect(next[0].itemId).toBe('fc1')
    expect(next[1].itemId).toBe('g1')
  })

  it('reorderRootItems returns same array for invalid ids', () => {
    const items = defaultRootItems(tree)
    expect(reorderRootItems(items, 'missing', 'g1')).toBe(items)
    expect(reorderRootItems(items, 'g1', 'missing')).toBe(items)
  })

  it('reorderMemberItems swaps group members', () => {
    const members = defaultMemberItems(tree.groups[0])
    const next = reorderMemberItems(members, 'g2', 'c1')
    expect(next[0].memberId).toBe('g2')
    expect(next[1].memberId).toBe('c1')
  })

  it('reorderMemberItems returns same array for invalid ids', () => {
    const members = defaultMemberItems(tree.groups[0])
    expect(reorderMemberItems(members, 'nope', 'c1')).toBe(members)
  })
})
