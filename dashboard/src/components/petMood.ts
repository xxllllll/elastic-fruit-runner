import type { TranslationKey } from '../i18n'

export type PetMood = 'idle' | 'busy' | 'sleeping' | 'alert'

export const MOOD_TRANSLATIONS: Record<PetMood, { label: TranslationKey; subtext: TranslationKey }> = {
  idle:     { label: 'mood.idle.label', subtext: 'mood.idle.subtext' },
  busy:     { label: 'mood.busy.label', subtext: 'mood.busy.subtext' },
  sleeping: { label: 'mood.sleeping.label', subtext: 'mood.sleeping.subtext' },
  alert:    { label: 'mood.alert.label', subtext: 'mood.alert.subtext' },
}
