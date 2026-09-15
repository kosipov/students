import { describe, expect, it } from 'vitest'
import type { Presence, ScheduleLesson } from '../api/types'
import { describePresence, formatLessonTime, lessonPlace } from './presence'

const lecture: ScheduleLesson = {
  number: 6,
  title: 'Проектирование и разработка веб-приложений',
  kind: 'л.',
  location: '245/1',
  room: '245',
  building: '1',
  groups: '1435-ИРо',
  startsAt: '2026-09-15T14:40:00Z',
  endsAt: '2026-09-15T16:10:00Z',
}
const practice: ScheduleLesson = { ...lecture, number: 7, kind: 'пр.', startsAt: '2026-09-15T16:20:00Z', endsAt: '2026-09-15T17:50:00Z' }
const presence = (patch: Partial<Presence>): Presence => ({ status: 'away', current: null, next: null, syncedAt: '2026-09-15T12:00:00Z', stale: false, ...patch })

describe('presence text', () => {
  it('shows lesson times in Moscow time', () => {
    expect(formatLessonTime(lecture.startsAt)).toBe('17:40')
  })

  it('names the room and building', () => {
    expect(lessonPlace(lecture)).toBe('ауд. 245, корпус 1')
    expect(lessonPlace({ ...lecture, room: '', building: '', location: 'дистанционно' })).toBe('дистанционно')
  })

  it('describes a lesson going on', () => {
    expect(describePresence(presence({ status: 'in_class', current: lecture, next: practice }))).toEqual({
      title: 'Сейчас я на паре — ауд. 245, корпус 1',
      details: ['Проектирование и разработка веб-приложений (лекция) · 1435-ИРо', 'до 19:10'],
      tone: 'here',
    })
  })

  it('mentions the next room when it changes', () => {
    const text = describePresence(presence({ status: 'in_class', current: lecture, next: { ...practice, room: '241', location: '241/1' } }))
    expect(text?.details).toContain('потом 19:20 — ауд. 241, корпус 1')
  })

  it('describes a break and being away', () => {
    expect(describePresence(presence({ status: 'break', next: practice }))?.title).toBe('Сейчас перерыв, я в университете')
    expect(describePresence(presence({ status: 'away', next: lecture }))).toEqual({
      title: 'Сейчас меня в университете нет',
      details: ['Сегодня буду с 17:40 — ауд. 245, корпус 1'],
      tone: 'away',
    })
    expect(describePresence(presence({ status: 'away' }))?.details).toEqual([])
  })

  it('says nothing before the schedule is loaded', () => {
    expect(describePresence(presence({ status: 'unknown', syncedAt: null }))).toBeNull()
  })
})
