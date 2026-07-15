import type { Runner, RunnerSet } from '../types'
import { elapsed, fmtDuration, shortName } from '../utils'
import type { TranslationKey } from '../i18n'
import { useI18n } from '../hooks/useI18n'

function StateLabel({ state }: { state: Runner['state'] }) {
  const { t } = useI18n()
  const map: Record<Runner['state'], { label: TranslationKey; color: string }> = {
    busy:      { label: 'runner.busy', color: 'var(--text)' },
    idle:      { label: 'runner.idle', color: 'var(--state-idle)' },
    preparing: { label: 'runner.preparing', color: 'var(--state-preparing)' },
    unknown:   { label: 'runner.unknown', color: 'var(--state-unknown)' },
  }
  const { label, color } = map[state]
  return (
    <span className="badge" style={{ color, borderColor: color }}>
      {t(label)}
    </span>
  )
}

function RunnerRow({ runner, now }: { runner: Runner; now: Date }) {
  const sec = elapsed(runner.since, now)
  return (
    <div className="runner-row">
      <span style={{ color: 'var(--runner-name)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
        {shortName(runner.name)}
      </span>
      <StateLabel state={runner.state} />
      <span style={{ color: 'var(--text-secondary)', textAlign: 'right', fontVariantNumeric: 'tabular-nums' }}>
        {fmtDuration(sec)}
      </span>
    </div>
  )
}

export function RunnerSetPanel({ rs, now }: { rs: RunnerSet; now: Date }) {
  const { t } = useI18n()
  const count = rs.runners.length
  const util = count / rs.maxRunners
  const busyCount    = rs.runners.filter(r => r.state === 'busy').length
  const idleCount    = rs.runners.filter(r => r.state === 'idle').length
  const prepCount    = rs.runners.filter(r => r.state === 'preparing').length
  const unknownCount = rs.runners.filter(r => r.state === 'unknown').length

  return (
    <div style={{ marginBottom: 28 }}>
      {/* Set header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 6 }}>
        <span style={{ fontWeight: 700, fontSize: 13, letterSpacing: '0.04em' }}>
          {rs.name}
        </span>
        <span style={{ color: 'var(--text-tertiary)', fontSize: 10, letterSpacing: '0.12em' }}>
          {rs.backend.toUpperCase()} · {count}/{rs.maxRunners}
        </span>
      </div>

      {/* Image */}
      <div style={{ color: 'var(--text-tertiary)', fontSize: 10, marginBottom: 10, letterSpacing: '0.04em', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
        {rs.image}
      </div>

      {/* Progress bar */}
      <div style={{ marginBottom: 8 }}>
        <div className="progress-track">
          <div className="progress-fill" style={{ width: `${util * 100}%` }} />
        </div>
      </div>

      {/* State counts */}
      <div style={{ display: 'flex', gap: 16, marginBottom: 12, fontSize: 10, letterSpacing: '0.12em' }}>
        {busyCount > 0 && <span style={{ color: 'var(--text)' }}>{t('runner.count', { count: busyCount, state: t('runner.busy') })}</span>}
        {idleCount > 0 && <span style={{ color: 'var(--state-idle)' }}>{t('runner.count', { count: idleCount, state: t('runner.idle') })}</span>}
        {prepCount > 0 && <span style={{ color: 'var(--state-preparing)' }}>{t('runner.count', { count: prepCount, state: t('runner.preparing') })}</span>}
        {unknownCount > 0 && <span style={{ color: 'var(--state-unknown)' }}>{t('runner.count', { count: unknownCount, state: t('runner.unknown') })}</span>}
        {count === 0 && <span style={{ color: 'var(--text-muted)' }}>{t('runner.none')}</span>}
      </div>

      {/* Column headers */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 64px 64px', gap: 10, paddingBottom: 5, marginBottom: 4, borderBottom: '1px solid var(--border)', fontSize: 9, color: 'var(--text-muted)', letterSpacing: '0.15em', textTransform: 'uppercase' }}>
        <span>{t('table.runner')}</span>
        <span>{t('table.state')}</span>
        <span style={{ textAlign: 'right' }}>{t('table.duration')}</span>
      </div>

      {/* Runners */}
      {rs.runners.map(r => (
        <RunnerRow key={r.name} runner={r} now={now} />
      ))}
      {rs.runners.length === 0 && (
        <div style={{ color: 'var(--text-muted)', fontSize: 12, padding: '8px 0' }}>{t('runner.noneActive')}</div>
      )}
    </div>
  )
}
