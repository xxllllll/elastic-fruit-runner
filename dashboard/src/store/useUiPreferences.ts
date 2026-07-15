import { create } from 'zustand'
import type { Locale } from '../i18n'

export type Theme = 'dark' | 'light'

const LOCALE_KEY = 'efr.locale'
const THEME_KEY = 'efr.theme'

function readStorage(key: string): string | null {
  try {
    return window.localStorage.getItem(key)
  } catch {
    return null
  }
}

function writeStorage(key: string, value: string) {
  try {
    window.localStorage.setItem(key, value)
  } catch {
    // Browser privacy settings may disable localStorage. The in-memory choice still works.
  }
}

function detectLocale(): Locale {
  const saved = readStorage(LOCALE_KEY)
  if (saved === 'en' || saved === 'zh-CN') return saved
  return navigator.language.toLowerCase().startsWith('zh') ? 'zh-CN' : 'en'
}

function detectTheme(): { theme: Theme; followsSystem: boolean } {
  const saved = readStorage(THEME_KEY)
  if (saved === 'dark' || saved === 'light') return { theme: saved, followsSystem: false }
  const light = window.matchMedia('(prefers-color-scheme: light)').matches
  return { theme: light ? 'light' : 'dark', followsSystem: true }
}

const initialTheme = detectTheme()

interface UiPreferencesState {
  locale: Locale
  theme: Theme
  followsSystemTheme: boolean
  setLocale: (locale: Locale) => void
  toggleLocale: () => void
  setTheme: (theme: Theme) => void
  toggleTheme: () => void
}

export const useUiPreferences = create<UiPreferencesState>((set, get) => ({
  locale: detectLocale(),
  theme: initialTheme.theme,
  followsSystemTheme: initialTheme.followsSystem,
  setLocale: (locale) => {
    writeStorage(LOCALE_KEY, locale)
    set({ locale })
  },
  toggleLocale: () => get().setLocale(get().locale === 'en' ? 'zh-CN' : 'en'),
  setTheme: (theme) => {
    writeStorage(THEME_KEY, theme)
    set({ theme, followsSystemTheme: false })
  },
  toggleTheme: () => get().setTheme(get().theme === 'dark' ? 'light' : 'dark'),
}))

let initialized = false

function applyDocumentPreferences(locale: Locale, theme: Theme) {
  document.documentElement.lang = locale
  document.documentElement.dataset.theme = theme
  document.documentElement.style.colorScheme = theme
}

export function initializeUiPreferences() {
  if (initialized) return
  initialized = true

  const state = useUiPreferences.getState()
  applyDocumentPreferences(state.locale, state.theme)
  useUiPreferences.subscribe(({ locale, theme }) => applyDocumentPreferences(locale, theme))

  const media = window.matchMedia('(prefers-color-scheme: light)')
  media.addEventListener('change', (event) => {
    if (!useUiPreferences.getState().followsSystemTheme) return
    useUiPreferences.setState({ theme: event.matches ? 'light' : 'dark' })
  })
}
