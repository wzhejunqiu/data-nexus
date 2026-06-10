import { describe, expect, it } from 'vitest'
import { defaultConnectionFormState } from './connectionFormDefaults'
import {
  connectionFailureFieldErrors,
  formatFieldError,
  parseConnectionFailureReason,
  validateConnectionField,
  validateConnectionForm,
} from './connectionFormValidation'

const t = (key: string, opts?: { field?: string }) => (opts?.field ? `${key}:${opts.field}` : key)

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

describe('validateConnectionField', () => {
  it('validates host on blur without other fields', () => {
    const state = {
      ...defaultConnectionFormState('postgres'),
      postgres: { ...defaultConnectionFormState('postgres').postgres, host: '' },
    }
    expect(validateConnectionField('host', state, 'create', 'postgres')).toBe('required')
  })

  it('validates port range on blur', () => {
    const state = {
      ...defaultConnectionFormState('postgres'),
      postgres: { ...defaultConnectionFormState('postgres').postgres, port: 0 },
    }
    expect(validateConnectionField('port', state, 'create', 'postgres')).toBe('invalidPort')
  })

  it('skips mysql database on blur', () => {
    const state = {
      ...defaultConnectionFormState('mysql'),
      mysql: { ...defaultConnectionFormState('mysql').mysql, database: '' },
    }
    expect(validateConnectionField('database', state, 'create', 'mysql')).toBeUndefined()
  })
})

describe('connectionFailureFieldErrors', () => {
  it('maps auth to user and password', () => {
    expect(connectionFailureFieldErrors('auth')).toEqual({ user: 'auth', password: 'auth' })
  })

  it('maps network to host and port', () => {
    expect(connectionFailureFieldErrors('network')).toEqual({ host: 'network', port: 'network' })
  })

  it('maps database to database field', () => {
    expect(connectionFailureFieldErrors('database')).toEqual({ database: 'database' })
  })
})

describe('parseConnectionFailureReason', () => {
  it('extracts reason from CONNECTION_FAILED', () => {
    expect(
      parseConnectionFailureReason({
        code: 'CONNECTION_FAILED',
        details: { reason: 'auth' },
      }),
    ).toBe('auth')
  })

  it('returns null for other errors', () => {
    expect(parseConnectionFailureReason({ code: 'INTERNAL' })).toBeNull()
  })
})

describe('formatFieldError', () => {
  it('formats connection failure codes', () => {
    expect(formatFieldError(t, 'user', 'auth')).toBe('connection.error.auth')
    expect(formatFieldError(t, 'host', 'network')).toBe('connection.error.network')
    expect(formatFieldError(t, 'database', 'database')).toBe('connection.error.database')
  })
})
