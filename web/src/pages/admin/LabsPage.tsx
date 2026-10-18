import { useState } from 'react'
import { useLabsOccupancy } from '../../api/queries'
import type { LabsFloor, LabRoom } from '../../api/types'
import { Corners } from '../../components/Corners'
import { Loading, QueryError } from '../../components/PageState'
import { useDocumentTitle } from '../../lib/hooks'
import { formatLessonTime, formatSyncTime, lessonKind } from '../../lib/presence'

const FLOORS: { key: LabsFloor; label: string }[] = [
  { key: '2', label: '2 этаж' },
  { key: '5', label: '5 этаж' },
  { key: 'all', label: 'Все' },
]

/** The university's timetable: when each pair starts and ends, Moscow time. */
const PAIR_TIMES = ['08:30–10:00', '10:10–11:40', '12:40–14:10', '14:20–15:50', '16:00–17:30', '17:40–19:10', '19:20–20:50', '21:00–22:30']

/** "245/1" → "245": the building is always 1, the number is what matters. */
function roomNumber(room: string): string {
  return room.split('/')[0] ?? room
}

function RoomRow({ room }: { room: LabRoom }) {
  const lesson = room.lesson
  return (
    <div className="panel-row">
      <span>
        <span className="panel-row-title">
          {roomNumber(room.room)}{' '}
          {room.failed ? (
            <span className="tag tag-neutral">Нет данных</span>
          ) : room.busy ? (
            <span className="tag tag-busy">Занята</span>
          ) : (
            <span className="tag tag-free">Свободна</span>
          )}
        </span>
        {lesson && (
          <>
            <span className="panel-row-sub">
              {lesson.title}
              {lesson.kind ? ` (${lessonKind(lesson.kind)})` : ''}
            </span>
            <span className="panel-row-sub">
              {lesson.teacher}
              {lesson.teacher && lesson.groups ? ' · ' : ''}
              {lesson.groups}
            </span>
          </>
        )}
        {room.failed && <span className="panel-row-problem">Кампус не ответил — попробуйте позже</span>}
        {room.stale && !room.failed && <span className="panel-row-problem">Кампус не ответил — показаны данные, которые могли устареть</span>}
      </span>
    </div>
  )
}

export function LabsPage() {
  useDocumentTitle('Лаборатории')
  const [floor, setFloor] = useState<LabsFloor | null>(null)
  const [pair, setPair] = useState<string | null>(null)
  const occupancy = useLabsOccupancy(floor, pair)

  const floorLabel = FLOORS.find((f) => f.key === floor)?.label ?? ''
  const data = occupancy.data

  return (
    <div>
      <h1 className="page-title ri">Лаборатории</h1>

      <section className="blueprint panel ri" aria-labelledby="labs-floor-title">
        <Corners />
        <h3 id="labs-floor-title">Этаж</h3>
        <div className="chips">
          {FLOORS.map((f) => (
            <button
              key={f.key}
              type="button"
              className={`btn ${floor === f.key ? 'btn-primary' : 'btn-secondary'}`}
              onClick={() => setFloor(f.key)}
            >
              {f.label}
            </button>
          ))}
        </div>
      </section>

      {floor !== null && (
        <section className="blueprint panel ri delay-1" aria-labelledby="labs-pair-title">
          <Corners />
          <h3 id="labs-pair-title">Пара</h3>
          <div className="chips">
            <button
              type="button"
              className={`btn ${pair === 'now' ? 'btn-primary' : 'btn-secondary'}`}
              onClick={() => setPair('now')}
            >
              Сейчас
            </button>
            {PAIR_TIMES.map((time, index) => (
              <button
                key={time}
                type="button"
                title={time}
                className={`btn ${pair === String(index + 1) ? 'btn-primary' : 'btn-secondary'}`}
                onClick={() => setPair(String(index + 1))}
              >
                {index + 1}
              </button>
            ))}
          </div>
          <p className="muted">1 — 08:30–10:00 · 2 — 10:10–11:40 · 3 — 12:40–14:10 · 4 — 14:20–15:50 · 5 — 16:00–17:30 · 6 — 17:40–19:10 · 7 — 19:20–20:50 · 8 — 21:00–22:30</p>
        </section>
      )}

      {floor !== null && pair !== null && (
        <section className="blueprint panel ri delay-2" aria-labelledby="labs-result-title">
          <Corners />
          <h3 id="labs-result-title">Занятость</h3>
          {occupancy.isPending && <Loading />}
          {occupancy.isError && <QueryError error={occupancy.error} retry={() => occupancy.refetch()} />}
          {data && (
            <>
              <p className="muted">
                {floorLabel} · {data.pair}-я пара, {formatLessonTime(data.startsAt)}–{formatLessonTime(data.endsAt)}
                {' · '}
                {new Date(data.date).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', timeZone: 'Europe/Moscow' })}
              </p>
              <div>
                {data.rooms.map((room) => (
                  <RoomRow key={room.room} room={room} />
                ))}
              </div>
              <p className="muted">
                {data.fetchedAt
                  ? `Данные campus.syktsu.ru от ${formatSyncTime(data.fetchedAt)} · обновляются не чаще раза в 12 часов`
                  : 'Кампус не ответил, данных нет'}
              </p>
            </>
          )}
        </section>
      )}
    </div>
  )
}
