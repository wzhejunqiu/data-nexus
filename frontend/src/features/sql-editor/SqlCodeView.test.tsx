import { describe, expect, it, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/render'
import { SqlCodeView } from './SqlCodeView'

vi.mock('@monaco-editor/react', () => ({
  default: ({ value }: { value: string }) => <div data-testid="monaco-editor">{value}</div>,
}))

describe('SqlCodeView', () => {
  it('renders formatted sql in monaco', () => {
    renderWithProviders(
      <SqlCodeView sql={'CREATE TABLE "t" ("id" INTEGER PRIMARY KEY)'} dialect="sqlite" />,
    )
    expect(screen.getByTestId('sql-code-view')).toBeTruthy()
    expect(screen.getByTestId('monaco-editor').textContent).toContain('CREATE TABLE')
  })
})
