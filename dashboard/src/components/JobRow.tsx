import type { JobRecord } from '../types'
import { elapsed, fmtDuration, shortName } from '../utils'
import { useI18n } from '../hooks/useI18n'

export function JobRow({ job, now }: { job: JobRecord; now: Date }) {
  const { t } = useI18n()
  const isRunning = job.result === 'running'
  const duration = job.completedAt
    ? Math.floor((job.completedAt.getTime() - job.startedAt.getTime()) / 1000)
    : elapsed(job.startedAt, now)
  const runnerLabel = job.runnerName ? shortName(job.runnerName) : t('common.unknown')

  const resultEl = isRunning
    ? <span style={{ color: 'var(--state-preparing)', fontSize: 10, letterSpacing: '0.08em' }} className="pulse">{t('job.running')}</span>
    : job.result === 'success'
    ? <span style={{ color: 'var(--text)', fontSize: 10, letterSpacing: '0.08em' }}>{t('job.success')}</span>
    : job.result === 'canceled'
    ? <span style={{ color: 'var(--warn)', fontSize: 10, letterSpacing: '0.08em' }}>{t('job.canceled')}</span>
    : <span style={{ color: 'var(--danger)', fontSize: 10, letterSpacing: '0.08em' }}>{t('job.failure')}</span>

  return (
    <div className="job-row" style={{ opacity: isRunning ? 1 : 0.7 }}>
      <span style={{
        color: isRunning ? 'var(--text)' : 'var(--text-secondary)',
        overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
      }}>
        {runnerLabel}
      </span>
      {resultEl}
      <span style={{ color: 'var(--text-tertiary)', textAlign: 'right', fontVariantNumeric: 'tabular-nums' }}>
        {fmtDuration(duration)}
      </span>
    </div>
  )
}
