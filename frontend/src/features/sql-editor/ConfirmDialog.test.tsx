import { describe, expect, it, vi } from 'vitest'
import { screen, fireEvent } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { ConfirmDialog } from './ConfirmDialog'

describe('ConfirmDialog', () => {
  it('calls onConfirm when confirmed', () => {
    const onConfirm = vi.fn()
    const onOpenChange = vi.fn()
    renderWithProviders(
      <ConfirmDialog
        open
        message="confirm write"
        onOpenChange={onOpenChange}
        onConfirm={onConfirm}
      />,
    )
    fireEvent.click(screen.getByRole('button', { name: /确认/i }))
    expect(onConfirm).toHaveBeenCalled()
  })

  it('calls onOpenChange(false) when cancelled', () => {
    const onOpenChange = vi.fn()
    renderWithProviders(
      <ConfirmDialog
        open
        message="confirm write"
        onOpenChange={onOpenChange}
        onConfirm={vi.fn()}
      />,
    )
    fireEvent.click(screen.getByRole('button', { name: /取消/i }))
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })
})
