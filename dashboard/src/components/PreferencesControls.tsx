import { useI18n } from '../hooks/useI18n'
import { useUiPreferences } from '../store/useUiPreferences'

export function PreferencesControls() {
  const { locale, t } = useI18n()
  const theme = useUiPreferences(state => state.theme)
  const toggleLocale = useUiPreferences(state => state.toggleLocale)
  const toggleTheme = useUiPreferences(state => state.toggleTheme)

  const languageLabel = locale === 'en' ? 'EN' : '简中'
  const languageAria = locale === 'en'
    ? t('preferences.switchLanguage')
    : t('preferences.switchLanguageZh')
  const themeLabel = theme === 'dark' ? t('preferences.dark') : t('preferences.light')
  const themeAria = theme === 'dark'
    ? t('preferences.switchThemeLight')
    : t('preferences.switchThemeDark')

  return (
    <div className="preference-controls">
      <button type="button" className="preference-button" onClick={toggleLocale} aria-label={languageAria}>
        <span className="preference-key">{t('preferences.language')}</span>
        <span>{languageLabel}</span>
      </button>
      <button type="button" className="preference-button" onClick={toggleTheme} aria-label={themeAria}>
        <span className="preference-key">{t('preferences.theme')}</span>
        <span>{themeLabel}</span>
      </button>
    </div>
  )
}
