import { usePresence } from '../api/queries'
import { describePresence, formatSyncTime } from '../lib/presence'
import { Corners } from './Corners'

/** "Where to find me now" from the university's schedule. Hidden while unknown or when it fails to load. */
export function PresenceCard() {
  const presence = usePresence()
  if (!presence.data) return null

  const text = describePresence(presence.data)
  if (!text) return null

  return (
    <section className={`blueprint presence presence--${text.tone} ri`} aria-live="polite" aria-labelledby="presence-title">
      <Corners />
      <span className="presence-dot" aria-hidden="true" />
      <div>
        <div className="presence-kicker">По расписанию университета</div>
        <h2 id="presence-title" className="presence-title">
          {text.title}
        </h2>
        {text.details.map((line) => (
          <p key={line} className="presence-details">
            {line}
          </p>
        ))}
        {presence.data.stale && presence.data.syncedAt && (
          <p className="presence-stale">
            Расписание не удавалось обновить с {formatSyncTime(presence.data.syncedAt)} — оно могло измениться.
          </p>
        )}
      </div>
    </section>
  )
}
