import { useI18n } from '../hooks/useI18n'

export function ConnectionStatus({ connected }: { connected: boolean | null }) {
  const { t } = useI18n()

  if (connected === null) {
    return (
      <>
        <span className="pulse" style={{ fontSize: 10, color: 'var(--text-secondary)' }}>●</span>
        <span style={{ fontSize: 11, letterSpacing: '0.12em', color: 'var(--text-secondary)' }}>
          {t('connection.checking')}
        </span>
      </>
    )
  }

  const color = connected ? 'var(--text)' : 'var(--danger)'
  const label = connected ? t('connection.connected') : t('connection.disconnected')

  return (
    <>
      <span className="pulse" style={{ fontSize: 10, color }}>●</span>
      <span style={{ fontSize: 11, letterSpacing: '0.12em', color }}>
        {label}
      </span>
    </>
  )
}
