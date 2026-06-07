import { describe, expect, it, vi } from 'vitest'
import { screen, fireEvent } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { IndexList, RenameDialog } from './IndexList'

describe('IndexList', () => {
  it('returns null for empty indexes', () => {
    const { container } = renderWithProviders(<IndexList indexes={[]} />)
    expect(container).toBeEmptyDOMElement()
  })

  it('renders unique marker', () => {
    renderWithProviders(
      <IndexList indexes={[{ name: 'idx_email', columns: ['email'], unique: true, primary: false }]} />,
    )
    expect(screen.getByText(/idx_email/)).toBeInTheDocument()
    expect(screen.getByText(/\[unique\]/)).toBeInTheDocument()
  })
})

describe('RenameDialog', () => {
  it('updates name and confirms', () => {
    const onNameChange = vi.fn()
    const onConfirm = vi.fn()
    renderWithProviders(
      <RenameDialog
        open
        name="old.db"
        onNameChange={onNameChange}
        onOpenChange={vi.fn()}
        onConfirm={onConfirm}
      />,
    )
    fireEvent.change(screen.getByDisplayValue('old.db'), { target: { value: 'new.db' } })
    expect(onNameChange).toHaveBeenCalledWith('new.db')
    fireEvent.click(screen.getByRole('button', { name: /确认/i }))
    expect(onConfirm).toHaveBeenCalled()
  })
})
