import { describe, expect, it, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { QueryHistory } from './QueryHistory'

describe('QueryHistory', () => {
  it('returns null when history is empty', () => {
    const { container } = render(<QueryHistory history={[]} onSelect={vi.fn()} />)
    expect(container).toBeEmptyDOMElement()
  })

  it('calls onSelect when an item is chosen', () => {
    const onSelect = vi.fn()
    render(<QueryHistory history={['SELECT 1', 'SELECT 2']} onSelect={onSelect} />)
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'SELECT 2' } })
    expect(onSelect).toHaveBeenCalledWith('SELECT 2')
  })
})
