import { describe, expect, it } from 'vitest'
import { filtersToSQL } from './filtersToSQL'

describe('filtersToSQL', () => {
  it('builds eq filter', () => {
    const sql = filtersToSQL('users', [{ column: 'name', operator: 'eq', value: "O'Brien" }])
    expect(sql).toBe(`SELECT * FROM "users" WHERE "name" = 'O''Brien';`)
  })

  it('wraps like without wildcards', () => {
    const sql = filtersToSQL('users', [{ column: 'email', operator: 'like', value: 'test' }])
    expect(sql).toBe(`SELECT * FROM "users" WHERE "email" LIKE '%test%';`)
  })

  it('includes order by', () => {
    const sql = filtersToSQL('users', [], 'id', 'desc')
    expect(sql).toBe('SELECT * FROM "users" ORDER BY "id" DESC;')
  })
})
