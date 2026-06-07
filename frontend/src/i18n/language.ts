import i18n from './index'

export const APP_LANGUAGES = ['zh-CN', 'en'] as const
export type AppLanguage = (typeof APP_LANGUAGES)[number]

export function isAppLanguage(value: string): value is AppLanguage {
  return (APP_LANGUAGES as readonly string[]).includes(value)
}

export function getAppLanguage(): AppLanguage {
  const fromI18n = i18n.language
  if (isAppLanguage(fromI18n)) return fromI18n
  if (typeof localStorage !== 'undefined') {
    const saved = localStorage.getItem('data-nexus-lang')
    if (saved && isAppLanguage(saved)) return saved
  }
  return 'zh-CN'
}

export function setAppLanguage(lang: AppLanguage) {
  void i18n.changeLanguage(lang)
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('data-nexus-lang', lang)
  }
}
