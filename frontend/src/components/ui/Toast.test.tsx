import { describe, expect, it, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Toaster, useToastStore } from './Toast'

describe('Toast store', () => {
  beforeEach(() => {
    useToastStore.setState({ items: [] })
    vi.useFakeTimers()
  })

  it('push adds toast item', () => {
    useToastStore.getState().push('hello', 'info')
    expect(useToastStore.getState().items).toHaveLength(1)
    expect(useToastStore.getState().items[0].message).toBe('hello')
  })

  it('dismiss removes toast item', () => {
    useToastStore.getState().push('bye', 'error')
    const id = useToastStore.getState().items[0].id
    useToastStore.getState().dismiss(id)
    expect(useToastStore.getState().items).toHaveLength(0)
  })
})

describe('Toaster', () => {
  beforeEach(() => {
    useToastStore.setState({ items: [] })
  })

  it('renders nothing when empty', () => {
    const { container } = render(<Toaster />)
    expect(container).toBeEmptyDOMElement()
  })

  it('renders error and info variants', () => {
    useToastStore.setState({
      items: [
        { id: '1', message: 'err', variant: 'error' },
        { id: '2', message: 'info', variant: 'info' },
      ],
    })
    render(<Toaster />)
    expect(screen.getByText('err')).toBeInTheDocument()
    expect(screen.getByText('info')).toBeInTheDocument()
    fireEvent.click(screen.getByText('err'))
    expect(useToastStore.getState().items).toHaveLength(1)
  })
})
