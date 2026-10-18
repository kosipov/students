import { describe, expect, it } from 'vitest'
import { addDays, mondayOf, mskToday, campusDateRange } from './labs'

describe('mskToday', () => {
  it('is a calendar day', () => {
    expect(mskToday()).toMatch(/^\d{4}-\d{2}-\d{2}$/)
  })
})

describe('mondayOf', () => {
  it('returns Monday of the week', () => {
    expect(mondayOf('2026-10-19')).toBe('2026-10-19') // Monday itself
    expect(mondayOf('2026-10-25')).toBe('2026-10-19') // Sunday belongs to the same week
    expect(mondayOf('2026-10-21')).toBe('2026-10-19')
  })
})

describe('addDays', () => {
  it('crosses month boundaries', () => {
    expect(addDays('2026-10-19', 13)).toBe('2026-11-01')
    expect(addDays('2026-10-31', 1)).toBe('2026-11-01')
  })
})

describe('campusDateRange', () => {
  it('starts on a Monday and ends on the Sunday three weeks later', () => {
    const { min, max } = campusDateRange()
    expect(min).toBe(mondayOf(mskToday()))
    expect(max).toBe(addDays(min, 27))
  })
})
