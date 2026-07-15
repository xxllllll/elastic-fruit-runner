import { useDashboardSync } from './hooks/useDashboardSync'
import { useDashboardDerived } from './hooks/useDashboardDerived'
import { elapsed, fmtDuration, fmtUptime } from './utils'
import { PixelPet } from './components/PixelPet'
import { MOOD_TRANSLATIONS } from './components/petMood'
import { SystemVitals } from './components/SystemVitals'
import { RunnerSetPanel } from './components/RunnerSetPanel'
import { JobRow } from './components/JobRow'
import { ConnectionStatus } from './components/ConnectionStatus'
import { PreferencesControls } from './components/PreferencesControls'
import { useI18n } from './hooks/useI18n'

function formatBuildVersion(version: string, unknownLabel: string) {
  if (!version) return unknownLabel
  if (version === 'unknown') return unknownLabel
  if (version === 'dev' || version.startsWith('v')) return version
  return `v${version}`
}

export default function App() {
  const { t } = useI18n()
  const { isLoading, error } = useDashboardSync()
  const {
    daemonStatus,
    runnerSets,
    recentJobs,
    machineVitals,
    now,
    uptime,
    totalMax,
    totalActive,
    preparing,
    idle,
    busy,
    utilPct,
    successCount,
    failureCount,
    canceledCount,
    mood,
  } = useDashboardDerived()

  if (error) {
    return (
      <div className="app-container" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', flexDirection: 'column', gap: 12 }}>
        <span style={{ fontSize: 13, fontWeight: 700, letterSpacing: '0.06em', color: 'var(--danger)' }}>{t('app.failedToLoad')}</span>
        <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>{String(error)}</span>
      </div>
    )
  }

  if (isLoading || !daemonStatus) {
    return (
      <div className="app-container" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
        <div className="spinner" />
      </div>
    )
  }

  const version = daemonStatus.buildInfo?.main?.version && daemonStatus.buildInfo.main.version !== '(devel)'
    ? daemonStatus.buildInfo.main.version
    : 'dev'
  const vcsRevision = daemonStatus.buildInfo?.settings.find(setting => setting.key === 'vcs.revision')?.value ?? t('common.unknown')

  return (
    <div className="app-container">

      {/* ── Header ── */}
      <div className="app-header">
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 20 }}>
          <span style={{ fontSize: 15, fontWeight: 700, letterSpacing: '0.06em' }}>
            ELASTIC-FRUIT-RUNNER
          </span>
          <span style={{ color: 'var(--text-tertiary)', fontSize: 11, letterSpacing: '0.08em' }}>
            {formatBuildVersion(version, t('common.unknown'))} · {vcsRevision}
          </span>
        </div>
        <div className="header-actions">
          <PreferencesControls />
          <ConnectionStatus connected={daemonStatus.githubConnected} />
        </div>
      </div>

      {/* ── Top status grid ── */}
      <div className="grid-top">

        {/* Cell: Daemon status */}
        <div className="cell">
          <div className="label">{t('app.status')}</div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 16 }}>
            <div className="spinner" />
            <div>
              <div style={{ fontSize: 13, fontWeight: 700, letterSpacing: '0.05em', marginBottom: 4, color: 'var(--text)' }}>
                {t('app.listening')}
              </div>
              <div style={{ fontSize: 9, color: 'var(--text-tertiary)', letterSpacing: '0.1em' }}>
                {t('app.githubApi')}
              </div>
            </div>
          </div>

          <div style={{ borderTop: '1px solid var(--border-subtle)', paddingTop: 14, display: 'flex', flexDirection: 'column', gap: 10 }}>
            {/* Scope */}
            <div>
              <div style={{ fontSize: 9, color: 'var(--text-muted)', letterSpacing: '0.15em', marginBottom: 3 }}>{t('app.connectedTo')}</div>
              {runnerSets.map(rs => (
                <div key={rs.name} style={{ fontSize: 11, color: 'var(--text-secondary)', letterSpacing: '0.04em' }}>
                  {rs.scope}
                </div>
              )).filter((_, i) => i === 0)}
            </div>

            {/* Auth + Sets */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
              <div>
                <div style={{ fontSize: 9, color: 'var(--text-muted)', letterSpacing: '0.15em', marginBottom: 3 }}>{t('app.authMode')}</div>
                <div style={{ fontSize: 11, color: 'var(--text-secondary)' }}>{t('app.githubApp')}</div>
              </div>
              <div>
                <div style={{ fontSize: 9, color: 'var(--text-muted)', letterSpacing: '0.15em', marginBottom: 3 }}>{t('app.runnerSets')}</div>
                <div style={{ fontSize: 18, fontWeight: 700, color: 'var(--text)', lineHeight: 1 }}>{runnerSets.length}</div>
              </div>
            </div>

            {/* Last activity */}
            <div>
              <div style={{ fontSize: 9, color: 'var(--text-muted)', letterSpacing: '0.15em', marginBottom: 3 }}>{t('app.lastJob')}</div>
              <div style={{ fontSize: 11, color: 'var(--text-secondary)' }}>
                {(() => {
                  const latest = recentJobs.find(j => j.completedAt)
                  if (!latest?.completedAt) return '—'
                  const s = elapsed(latest.completedAt, now)
                  return t('app.ago', { duration: fmtDuration(s) })
                })()}
              </div>
            </div>
          </div>
        </div>

        {/* Cell: Capacity */}
        <div className="cell">
          <div className="label">{t('app.runnerCapacity')}</div>

          {/* Overall bar */}
          <div className="progress-track" style={{ marginBottom: 8 }}>
            <div className="progress-fill" style={{ width: `${utilPct}%` }} />
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 4 }}>
            <span className="value-md">{totalActive}</span>
            <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>{t('app.ofMax', { max: totalMax })}</span>
          </div>
          <div style={{ fontSize: 10, color: 'var(--text-tertiary)', letterSpacing: '0.12em', marginBottom: 16 }}>
            {t('app.utilization', { percent: utilPct, free: totalMax - totalActive })}
          </div>

          {/* Per-set breakdown */}
          <div style={{ borderTop: '1px solid var(--border-subtle)', paddingTop: 14 }}>
            <div style={{ fontSize: 9, color: 'var(--text-muted)', letterSpacing: '0.15em', marginBottom: 10 }}>{t('app.perSet')}</div>
            {runnerSets.map(rs => {
              const pct = Math.round((rs.runners.length / rs.maxRunners) * 100)
              return (
                <div key={rs.name} style={{ marginBottom: 10 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <span style={{ fontSize: 10, color: 'var(--text-secondary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: '70%' }}>
                      {rs.name}
                    </span>
                    <span style={{ fontSize: 10, color: 'var(--text-tertiary)', flexShrink: 0 }}>
                      {rs.runners.length}/{rs.maxRunners}
                    </span>
                  </div>
                  <div style={{ height: 4, background: 'var(--track)', overflow: 'hidden' }}>
                    <div style={{
                      height: '100%', width: `${pct}%`,
                      background: pct >= 80 ? 'var(--warn)' : 'var(--accent-soft)',
                      transition: 'width 0.6s ease',
                      backgroundImage: 'repeating-linear-gradient(90deg, transparent 0px, transparent 3px, var(--stripe) 3px, var(--stripe) 4px)',
                    }} />
                  </div>
                </div>
              )
            })}
          </div>

          {/* Throughput stats */}
          <div style={{ borderTop: '1px solid var(--border-subtle)', paddingTop: 14, display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
            <div>
              <div style={{ fontSize: 9, color: 'var(--text-muted)', letterSpacing: '0.15em', marginBottom: 3 }}>{t('app.avgDuration')}</div>
              <div style={{ fontSize: 16, fontWeight: 700, color: 'var(--text)' }}>
                {(() => {
                  const done = recentJobs.filter(j => j.completedAt)
                  if (!done.length) return '—'
                  const avg = done.reduce((s, j) => s + (j.completedAt!.getTime() - j.startedAt.getTime()) / 1000, 0) / done.length
                  return fmtDuration(Math.round(avg))
                })()}
              </div>
            </div>
            <div>
              <div style={{ fontSize: 9, color: 'var(--text-muted)', letterSpacing: '0.15em', marginBottom: 3 }}>{t('app.successRate')}</div>
              <div style={{ fontSize: 16, fontWeight: 700, color: successCount + failureCount + canceledCount > 0 && (failureCount + canceledCount) / (successCount + failureCount + canceledCount) > 0.2 ? 'var(--warn)' : 'var(--text)' }}>
                {successCount + failureCount + canceledCount > 0
                  ? `${Math.round(successCount / (successCount + failureCount + canceledCount) * 100)}%`
                  : '—'}
              </div>
            </div>
          </div>
        </div>

        {/* Cell: BRAIN — pixel pet + vitals */}
        <div className="cell cell-brain">
          <div className="label">{t('app.runnerBrain')}</div>
          <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>
            {/* Pixel pet */}
            <div style={{ flexShrink: 0 }}>
              <PixelPet mood={mood} />
              <div style={{ marginTop: 6, fontSize: 8, letterSpacing: '0.1em', color: 'var(--text-tertiary)', textAlign: 'center' }}>
                {t(MOOD_TRANSLATIONS[mood].label)}
              </div>
            </div>
            {/* Status text + uptime */}
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontSize: 12, fontWeight: 700, color: 'var(--text)', marginBottom: 3, letterSpacing: '0.04em' }}>
                {t(MOOD_TRANSLATIONS[mood].label)}
              </div>
              <div style={{ fontSize: 9, color: 'var(--text-tertiary)', letterSpacing: '0.06em', marginBottom: 12 }}>
                {t(MOOD_TRANSLATIONS[mood].subtext)}
              </div>
              <div style={{ fontSize: 9, color: 'var(--text-muted)', letterSpacing: '0.12em', marginBottom: 2 }}>{t('app.uptime')}</div>
              <div style={{ fontSize: 16, fontWeight: 700, fontVariantNumeric: 'tabular-nums', letterSpacing: '0.02em', marginBottom: 10 }}>
                {fmtUptime(uptime)}
              </div>
              <div style={{ fontSize: 9, color: 'var(--text-faint)', letterSpacing: '0.1em' }}>
                {t('app.idleTimeout', { minutes: Math.floor(daemonStatus.idleTimeout / 60) })}
              </div>
            </div>
          </div>
          {/* System vitals */}
          <div className="brain-vitals" style={{ borderTop: '1px solid var(--border-subtle)', marginTop: 14, paddingTop: 14 }}>
            <div className="label" style={{ marginBottom: 10 }}>{t('app.systemVitals')}</div>
            <SystemVitals vitals={machineVitals} />
          </div>
        </div>
      </div>

      {/* ── Stats row ── */}
      <div className="grid-stats">
        {[
          { label: t('stats.preparing'), value: preparing,    color: 'var(--state-preparing)' },
          { label: t('stats.idle'),      value: idle,         color: 'var(--state-idle)' },
          { label: t('stats.busy'),      value: busy,         color: 'var(--text)' },
          { label: t('stats.succeeded'), value: successCount, color: 'var(--text)' },
          { label: t('stats.failed'),    value: failureCount, color: failureCount > 0 ? 'var(--danger)' : 'var(--text-muted)' },
          { label: t('stats.canceled'),  value: canceledCount, color: canceledCount > 0 ? 'var(--warn)' : 'var(--text-muted)' },
        ].map((stat) => (
          <div key={stat.label} className="cell">
            <div className="label">{stat.label}</div>
            <div className="value-md" style={{ color: stat.color }}>
              {String(stat.value).padStart(2, '0')}
            </div>
          </div>
        ))}
      </div>

      {/* ── Main area ── */}
      <div className="grid-main">

        {/* Runner Sets */}
        <div className="cell">
          <div className="label" style={{ marginBottom: 22 }}>{t('app.runnerSets')}</div>
          {runnerSets.map(rs => (
            <RunnerSetPanel key={rs.name} rs={rs} now={now} />
          ))}
        </div>

        {/* Recent Jobs */}
        <div className="cell">
          <div className="label" style={{ marginBottom: 14 }}>{t('app.recentJobs')}</div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 54px 58px', gap: 10, paddingBottom: 6, marginBottom: 4, borderBottom: '1px solid var(--border)', fontSize: 9, color: 'var(--text-muted)', letterSpacing: '0.15em', textTransform: 'uppercase' }}>
            <span>{t('table.runner')}</span>
            <span>{t('table.result')}</span>
            <span style={{ textAlign: 'right' }}>{t('table.duration')}</span>
          </div>
          {recentJobs.map(job => (
            <JobRow key={job.id} job={job} now={now} />
          ))}
        </div>
      </div>

      {/* ── Footer ── */}
      <div style={{ marginTop: 20, display: 'flex', justifyContent: 'space-between', color: 'var(--text-faint)', fontSize: 10, letterSpacing: '0.12em' }}>
        <span>{t('app.scope', { scope: runnerSets[0]?.scope.toUpperCase() ?? '—' })}</span>
        <span>{t('app.autoRefresh')}<span className="blink">_</span></span>
      </div>
    </div>
  )
}
