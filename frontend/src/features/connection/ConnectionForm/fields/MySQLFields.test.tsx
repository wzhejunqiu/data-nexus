import { describe, expect, it, vi } from 'vitest'
import { useState } from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { defaultConnectionFormState, type ConnectionFormState } from '../connectionFormDefaults'
import { MySQLAdvancedFields } from './MySQLFields'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

function Harness() {
  const [state, setState] = useState(defaultConnectionFormState('mysql'))
  const onChange = (patch: Partial<ConnectionFormState>) => {
    setState((prev) => ({
      ...prev,
      ...patch,
      mysql: patch.mysql ? { ...prev.mysql, ...patch.mysql } : prev.mysql,
    }))
  }
  return <MySQLAdvancedFields state={state} onChange={onChange} />
}

describe('MySQLAdvancedFields', () => {
  it('updates collation when charset changes', () => {
    render(<Harness />)

    fireEvent.change(screen.getByLabelText('connectionForm.charset'), {
      target: { value: 'gbk' },
    })

    const collationSelect = screen.getByLabelText('connectionForm.collation') as HTMLSelectElement
    expect(collationSelect.value).toBe('gbk_chinese_ci')
    expect(collationSelect.options.length).toBe(1)
  })
})
