import { describe, expect, it, vi, afterEach } from 'vitest'
import { cleanup, render, screen, fireEvent, within } from '@testing-library/react'
import { Pagination } from './Pagination'

const pagination = { page: 2, pageSize: 50, totalRows: 120, totalPages: 3 }

function renderPagination(props: {
  pagination: typeof pagination
  page: number
  onPageChange?: (page: number) => void
}) {
  const view = render(
    <Pagination
      pagination={props.pagination}
      page={props.page}
      onPageChange={props.onPageChange ?? (() => {})}
    />,
  )
  return within(view.container)
}

describe('Pagination', () => {
  afterEach(() => {
    cleanup()
  })

  it('disables first/prev on first page', () => {
    const onPageChange = vi.fn()
    const ui = renderPagination({
      pagination: { ...pagination, page: 1, totalPages: 3 },
      page: 1,
      onPageChange,
    })

    const buttons = ui.getAllByRole('button') as HTMLButtonElement[]
    expect(buttons[0].disabled).toBe(true)
    expect(buttons[1].disabled).toBe(true)
    expect(ui.getByText('1 / 3')).toBeTruthy()
  })

  it('disables next/last on last page', () => {
    const ui = renderPagination({
      pagination: { page: 3, pageSize: 50, totalRows: 120, totalPages: 3 },
      page: 3,
    })

    const buttons = ui.getAllByRole('button') as HTMLButtonElement[]
    expect(buttons[2].disabled).toBe(true)
    expect(buttons[3].disabled).toBe(true)
  })

  it('calls onPageChange for navigation buttons', () => {
    const onPageChange = vi.fn()
    const ui = renderPagination({ pagination, page: 2, onPageChange })

    const buttons = ui.getAllByRole('button')
    fireEvent.click(buttons[0])
    expect(onPageChange).toHaveBeenCalledWith(1)

    fireEvent.click(buttons[1])
    expect(onPageChange).toHaveBeenCalledWith(1)

    fireEvent.click(buttons[2])
    expect(onPageChange).toHaveBeenCalledWith(3)

    fireEvent.click(buttons[3])
    expect(onPageChange).toHaveBeenCalledWith(3)
  })
})
