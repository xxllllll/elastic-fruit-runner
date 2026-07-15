import { translate, type TranslationKey } from '../i18n'
import { useUiPreferences } from '../store/useUiPreferences'

type TranslationParams = Record<string, string | number>

export function useI18n() {
  const locale = useUiPreferences(state => state.locale)
  const t = (key: TranslationKey, params?: TranslationParams) => translate(locale, key, params)
  return { locale, t }
}
