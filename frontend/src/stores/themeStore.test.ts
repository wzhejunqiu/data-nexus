import { describe, expect, it, beforeEach, vi } from 'vitest'
import { applyThemeMode, resolveTheme, useThemeStore } from './themeStore'

describe('themeStore', () => {
  beforeEach(() => {
    useThemeStore.setState({ mode: 'system' })
    useThemeStore.persist.clearStorage()
    document.documentElement.classList.remove('dark')
  })

  it('setMode updates mode', () => {
    useThemeStore.getState().setMode('dark')
    expect(useThemeStore.getState().mode).toBe('dark')
  })

  it('applyThemeMode updates store and document class', () => {
    applyThemeMode('dark')
    expect(useThemeStore.getState().mode).toBe('dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)

    applyThemeMode('light')
    expect(useThemeStore.getState().mode).toBe('light')
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('resolveTheme returns explicit modes', () => {
    expect(resolveTheme('light')).toBe('light')
    expect(resolveTheme('dark')).toBe('dark')
  })

  it('resolveTheme follows system preference', () => {
    window.matchMedia = vi.fn().mockReturnValue({
      matches: true,
      media: '',
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    } as MediaQueryList)
    expect(resolveTheme('system')).toBe('dark')

    window.matchMedia = vi.fn().mockReturnValue({
      matches: false,
      media: '',
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    } as MediaQueryList)
    expect(resolveTheme('system')).toBe('light')
  })
})
