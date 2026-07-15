import type { MachineVitals } from '../types'
import type { TranslationKey } from '../i18n'
import { useI18n } from '../hooks/useI18n'

interface VitalConfig {
  key: string
  label: TranslationKey
  unit: string
  max: number
  warn: number
  crit: number
}

const VITAL_CONFIGS: VitalConfig[] = [
  { key: 'cpu',  label: 'vitals.processor',   unit: '%',   max: 100, warn: 70, crit: 90 },
  { key: 'mem',  label: 'vitals.memory',      unit: '%',   max: 100, warn: 80, crit: 95 },
  { key: 'disk', label: 'vitals.storage',     unit: '%',   max: 100, warn: 85, crit: 95 },
  { key: 'temp', label: 'vitals.temperature', unit: '°C',  max: 100, warn: 70, crit: 85 },
]

function getValue(vitals: MachineVitals, key: string): number {
  switch (key) {
    case 'cpu':  return vitals.cpuUsagePercent
    case 'mem':  return vitals.memoryUsagePercent
    case 'disk': return vitals.diskUsagePercent
    case 'temp': return vitals.temperatureCelsius
    default:     return 0
  }
}

function VitalBar({ config, value }: { config: VitalConfig; value: number }) {
  const { t } = useI18n()
  const rounded = Math.round(value)
  const pct = Math.min(100, Math.max(0, (value / config.max) * 100))

  const color =
    rounded >= config.crit ? 'var(--danger)' :
    rounded >= config.warn ? 'var(--warn)' :
    'var(--text-secondary)'

  const barColor =
    rounded >= config.crit ? 'var(--danger)' :
    rounded >= config.warn ? 'var(--warn)' :
    'var(--accent-soft)'

  return (
    <div style={{ marginBottom: 14 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 5 }}>
        <span style={{ fontSize: 9, letterSpacing: '0.15em', color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>
          {t(config.label)}
        </span>
        <span style={{ fontSize: 13, fontWeight: 700, color, fontVariantNumeric: 'tabular-nums' }}>
          {rounded}{config.unit}
        </span>
      </div>
      <div style={{ height: 4, background: 'var(--track)', position: 'relative', overflow: 'hidden' }}>
        <div
          style={{
            height: '100%',
            width: `${pct}%`,
            background: barColor,
            transition: 'width 1.5s ease, background 0.5s ease',
            position: 'relative',
          }}
        >
          <div style={{
            position: 'absolute', inset: 0,
            background: 'repeating-linear-gradient(90deg, transparent 0px, transparent 3px, var(--stripe) 3px, var(--stripe) 4px)',
          }} />
        </div>
      </div>
    </div>
  )
}

export function SystemVitals({ vitals }: { vitals: MachineVitals | null }) {
  const { t } = useI18n()

  if (!vitals) {
    return (
      <div style={{ fontSize: 10, color: 'var(--text-muted)', letterSpacing: '0.1em' }}>
        {t('common.loading')}
      </div>
    )
  }

  return (
    <div>
      {VITAL_CONFIGS.map(config => (
        <VitalBar key={config.key} config={config} value={getValue(vitals, config.key)} />
      ))}
    </div>
  )
}
