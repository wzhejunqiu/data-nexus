import { describe, expect, it, vi, beforeEach } from 'vitest'
import { createRef } from 'react'
import { act, fireEvent, render, screen } from '@testing-library/react'
import { ConnectionForm, type ConnectionFormHandle } from './ConnectionForm'
import type { ConnectionListItem } from '@/lib/types'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

const closedItem: ConnectionListItem = {
  id: 'c1',
  name: 'app.db',
  type: 'sqlite',
  config: {
    type: 'sqlite',
    sqlite: { filePath: '/tmp/app.db', readOnly: false, wal: false },
  },
  status: 'closed',
  lastUsedAt: '2026-01-01T00:00:00Z',
}

describe('ConnectionForm', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('requires remote fields for postgres create', () => {
    const onSubmit = vi.fn()
    const ref = createRef<ConnectionFormHandle>()
    render(
      <ConnectionForm
        ref={ref}
        mode="create"
        initialDialect="postgres"
        onSubmit={onSubmit}
        showTest
        onTest={vi.fn()}
      />,
    )

    ref.current?.submit()
    expect(onSubmit).not.toHaveBeenCalled()
    expect(screen.getByLabelText(/connectionForm.fields.displayName/i)).toBeInTheDocument()
  })

  it('allows edit submit with empty password', () => {
    const onSubmit = vi.fn()
    const ref = createRef<ConnectionFormHandle>()
    render(
      <ConnectionForm
        ref={ref}
        mode="edit"
        initialItem={closedItem}
        onSubmit={onSubmit}
        dialectLocked
      />,
    )

    ref.current?.submit()
    expect(onSubmit).toHaveBeenCalled()
  })

  it('shows test button for remote dialects', () => {
    render(
      <ConnectionForm
        mode="create"
        initialDialect="mysql"
        onSubmit={vi.fn()}
        showTest
        onTest={vi.fn()}
      />,
    )
    expect(screen.getByText('connectionForm.testConnection')).toBeInTheDocument()
  })

  it('validates mysql required fields on submit', () => {
    const onSubmit = vi.fn()
    const ref = createRef<ConnectionFormHandle>()
    render(
      <ConnectionForm
        ref={ref}
        mode="create"
        initialDialect="mysql"
        onSubmit={onSubmit}
        showTest
        onTest={vi.fn()}
      />,
    )

    fireEvent.change(screen.getByLabelText(/connectionForm.fields.displayName/i), {
      target: { value: 'mysql-local' },
    })
    ref.current?.submit()
    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('highlights user and password on auth connection failure', () => {
    const ref = createRef<ConnectionFormHandle>()
    render(
      <ConnectionForm
        ref={ref}
        mode="create"
        initialDialect="postgres"
        onSubmit={vi.fn()}
        showTest
        onTest={vi.fn()}
      />,
    )

    act(() => {
      ref.current?.highlightConnectionFailure('auth')
    })
    expect(screen.getAllByText('connection.error.auth')).toHaveLength(2)
    expect(screen.getByLabelText(/connectionForm.fields.user/i)).toHaveAttribute(
      'aria-describedby',
      'user-error',
    )
  })

  it('validates host on blur', () => {
    render(
      <ConnectionForm
        mode="create"
        initialDialect="postgres"
        onSubmit={vi.fn()}
        showTest
        onTest={vi.fn()}
      />,
    )

    const host = screen.getByLabelText(/connectionForm.fields.host/i)
    fireEvent.change(host, { target: { value: '' } })
    fireEvent.blur(host)
    expect(screen.getByText(/connectionForm.validation.required/)).toBeInTheDocument()
    expect(host).toHaveAttribute('aria-describedby', 'host-error')
  })
})
