import { describe, expect, it } from 'vitest'
import type { Catalog, ScheduleLesson } from '../api/types'
import { groupsWithLessonFirst, matchLesson } from './currentLesson'

const catalog: Catalog = {
  groups: [
    { id: 1, name: '1435-ИРо', subjects: [{ id: 10, name: 'Проектирование и разработка веб-приложений', tasks: [] }] },
    { id: 2, name: '122-МКо', subjects: [{ id: 20, name: 'Параллельное программирование', tasks: [] }] },
  ],
}

const lesson = (patch: Partial<ScheduleLesson>): ScheduleLesson => ({
  number: 1,
  title: 'Проектирование и разработка веб-приложений',
  kind: 'л.',
  location: '245/1',
  room: '245',
  building: '1',
  groups: '1435-ИРо',
  startsAt: '2026-09-15T14:40:00Z',
  endsAt: '2026-09-15T16:10:00Z',
  ...patch,
})

describe('matchLesson', () => {
  it('finds the group and the subject of the lesson', () => {
    expect(matchLesson(catalog, lesson({}))).toMatchObject({ group: { id: 1 }, subject: { id: 10 } })
  })

  it('ignores case, ё and spaces in names', () => {
    expect(matchLesson(catalog, lesson({ groups: ' 1435-иро ', title: 'ПРОЕКТИРОВАНИЕ И РАЗРАБОТКА ВЕБ-ПРИЛОЖЕНИЙ' }))).toMatchObject({ subject: { id: 10 } })
  })

  it('takes the group that has the subject when a lesson is for several groups', () => {
    expect(matchLesson(catalog, lesson({ groups: '122-МКо, 1435-ИРо' }))).toMatchObject({ group: { id: 1 }, subject: { id: 10 } })
    expect(matchLesson(catalog, lesson({ groups: '1435-ИРо, 122-МКо', title: 'Параллельное программирование' }))).toMatchObject({ group: { id: 2 } })
  })

  it('matches a shortened subject name', () => {
    expect(matchLesson(catalog, lesson({ title: 'Проектирование и разработка веб' }))).toMatchObject({ subject: { id: 10 } })
  })

  it('finds nothing for an unknown group or subject', () => {
    expect(matchLesson(catalog, lesson({ groups: '999-ХХо' }))).toBeNull()
    expect(matchLesson(catalog, lesson({ title: 'Физкультура' }))).toBeNull()
    expect(matchLesson(catalog, null)).toBeNull()
  })

  it('searches every group when the lesson has no group written', () => {
    expect(matchLesson(catalog, lesson({ groups: '' }))).toMatchObject({ subject: { id: 10 } })
  })
})

describe('groupsWithLessonFirst', () => {
  const match = matchLesson(catalog, lesson({ groups: '122-МКо', title: 'Параллельное программирование' }))

  it('puts the group with a lesson now first', () => {
    expect(groupsWithLessonFirst(catalog.groups, match).map((g) => g.id)).toEqual([2, 1])
  })

  it('keeps the order when there is no lesson or the group is not in the list', () => {
    expect(groupsWithLessonFirst(catalog.groups, null).map((g) => g.id)).toEqual([1, 2])
    expect(groupsWithLessonFirst([catalog.groups[0]!], match).map((g) => g.id)).toEqual([1])
  })
})
