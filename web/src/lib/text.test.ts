import { describe, expect, it } from 'vitest'
import { countLabel, normalizeForSearch, ordinal, plural, safeNextPath, TASK_FORMS } from './text'

describe('plural', () => {
  it.each([
    [0, 'заданий'],
    [1, 'задание'],
    [2, 'задания'],
    [4, 'задания'],
    [5, 'заданий'],
    [11, 'заданий'],
    [12, 'заданий'],
    [14, 'заданий'],
    [21, 'задание'],
    [22, 'задания'],
    [111, 'заданий'],
    [101, 'задание'],
  ])('%i → %s', (count, expected) => {
    expect(plural(count, TASK_FORMS)).toBe(expected)
  })

  it('builds a count label', () => {
    expect(countLabel(3, TASK_FORMS)).toBe('3 задания')
  })
})

describe('ordinal', () => {
  it('pads to two digits starting from one', () => {
    expect(ordinal(0)).toBe('01')
    expect(ordinal(11)).toBe('12')
  })
})

describe('normalizeForSearch', () => {
  it('ignores case, spaces and ё', () => {
    expect(normalizeForSearch('  Учёт ')).toBe(normalizeForSearch('учет'))
  })
})

describe('safeNextPath', () => {
  it('keeps paths inside the site', () => {
    expect(safeNextPath('/admin/groups?x=1', '/admin')).toBe('/admin/groups?x=1')
  })

  it.each([null, '', 'https://evil.example', '//evil.example', '/\\evil.example', 'admin'])('rejects %s', (next) => {
    expect(safeNextPath(next, '/admin')).toBe('/admin')
  })
})
