import { describe, expect, it } from 'vitest'
import type { Catalog } from '../api/types'
import { countTasks, searchTasks } from './catalog'

const catalog: Catalog = {
  groups: [
    {
      id: 1,
      name: '1435-ИРо',
      subjects: [
        {
          id: 10,
          name: 'Проектирование веб-приложений',
          tasks: [
            { id: 100, name: 'Темы курсовых работ', comment: '', href: 'https://1drv.ms/t/c/1', isDocument: true },
            { id: 101, name: 'Требования к записке', comment: '', href: 'https://example.com', isDocument: false },
          ],
        },
      ],
    },
    {
      id: 2,
      name: '112-МКО',
      subjects: [
        {
          id: 20,
          name: 'Поисковая оптимизация',
          tasks: [{ id: 200, name: 'Лабораторная работа №1', comment: '', href: '', isDocument: false }],
        },
      ],
    },
  ],
}

describe('searchTasks', () => {
  it('finds tasks by name in any group', () => {
    const results = searchTasks(catalog, 'лаборат')
    expect(results.map((r) => r.task.id)).toEqual([200])
    expect(results[0]?.group.name).toBe('112-МКО')
    expect(results[0]?.subject.name).toBe('Поисковая оптимизация')
  })

  it('finds all tasks of a matching subject', () => {
    expect(searchTasks(catalog, 'ВЕБ').map((r) => r.task.id)).toEqual([100, 101])
  })

  it('returns nothing for a blank query', () => {
    expect(searchTasks(catalog, '   ')).toEqual([])
  })
})

describe('countTasks', () => {
  it('counts tasks of all subjects', () => {
    expect(countTasks(catalog.groups[0]!)).toBe(2)
  })
})
