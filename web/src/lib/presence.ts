import type { Presence, ScheduleLesson } from '../api/types'

const moscowTime = new Intl.DateTimeFormat('ru-RU', { hour: '2-digit', minute: '2-digit', timeZone: 'Europe/Moscow' })
const moscowDateTime = new Intl.DateTimeFormat('ru-RU', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit', timeZone: 'Europe/Moscow' })

/** Lesson times are shown in the university's time, whatever the visitor's time zone is. */
export function formatLessonTime(iso: string): string {
  return moscowTime.format(new Date(iso))
}

export function formatSyncTime(iso: string): string {
  return moscowDateTime.format(new Date(iso))
}

const KINDS: Record<string, string> = { 'л.': 'лекция', 'пр.': 'практика', 'лаб.': 'лабораторная' }

export function lessonKind(kind: string): string {
  return KINDS[kind] ?? kind
}

/** "ауд. 245, корпус 1" for "245/1"; other locations as written. */
export function lessonPlace(lesson: ScheduleLesson): string {
  if (lesson.room && lesson.building) return `ауд. ${lesson.room}, корпус ${lesson.building}`
  return lesson.location || 'аудитория не указана'
}

export interface PresenceText {
  title: string
  details: string[]
  tone: 'here' | 'break' | 'away'
}

/** What the presence card says; null when there is nothing to say (the schedule was never loaded). */
export function describePresence(presence: Presence): PresenceText | null {
  const { status, current, next } = presence
  if (status === 'unknown') return null

  if (status === 'in_class' && current) {
    const details = [`${current.title}${current.kind ? ` (${lessonKind(current.kind)})` : ''}${current.groups ? ` · ${current.groups}` : ''}`, `до ${formatLessonTime(current.endsAt)}`]
    if (next && lessonPlace(next) !== lessonPlace(current)) {
      details.push(`потом ${formatLessonTime(next.startsAt)} — ${lessonPlace(next)}`)
    }
    return { title: `Сейчас я на паре — ${lessonPlace(current)}`, details, tone: 'here' }
  }

  if (status === 'break' && next) {
    return {
      title: 'Сейчас перерыв, я в университете',
      details: [`Следующая пара в ${formatLessonTime(next.startsAt)} — ${lessonPlace(next)}`],
      tone: 'break',
    }
  }

  return {
    title: 'Сейчас меня в университете нет',
    details: next ? [`Сегодня буду с ${formatLessonTime(next.startsAt)} — ${lessonPlace(next)}`] : [],
    tone: 'away',
  }
}
