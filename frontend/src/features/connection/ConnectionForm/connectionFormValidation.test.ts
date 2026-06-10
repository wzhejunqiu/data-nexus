import { describe, expect, it } from 'vitest'
import { defaultConnectionFormState } from './connectionFormDefaults'
import { validateConnectionForm } from './connectionFormValidation'

describe('validateConnectionForm', () => {
  it('allows sqlite create with empty filePath', () => {
    const state = {
      ...defaultConnectionFormState('sqlite'),
      filePath: '',
    }
    expect(validateConnectionForm(state, 'create', 'sqlite')).toEqual({})
  })

  it('allows mysql without database on create', () => {
    const state = {
      ...defaultConnectionFormState('mysql'),
      displayName: 'local-mysql',
      password: 'secret',
      mysql: {
        ...defaultConnectionFormState('mysql').mysql,
        user: 'root',
        database: '',
      },
    }
    expect(validateConnectionForm(state, 'create', 'mysql')).toEqual({})
  })

  it('requires postgres database on create', () => {
    const state = {
      ...defaultConnectionFormState('postgres'),
      displayName: 'local-pg',
      password: 'secret',
      postgres: {
        ...defaultConnectionFormState('postgres').postgres,
        user: 'admin',
        database: '',
      },
    }
    expect(validateConnectionForm(state, 'create', 'postgres')).toEqual({
      database: 'required',
    })
  })

  it('allows mysql without database on edit', () => {
    const state = {
      ...defaultConnectionFormState('mysql'),
      displayName: 'local-mysql',
      mysql: {
        ...defaultConnectionFormState('mysql').mysql,
        user: 'root',
        database: '',
      },
    }
    expect(validateConnectionForm(state, 'edit', 'mysql')).toEqual({})
  })
})
