import { describe, expect, it } from 'vitest'
import { generatePassword } from './password'

describe('generatePassword', () => {
  it('makes readable passwords without confusable characters', () => {
    for (let i = 0; i < 200; i++) {
      expect(generatePassword()).toMatch(/^[abcdefghjkmnpqrstuvwxyz23456789]{8}$/)
    }
  })

  it('skips bytes that would make some characters more likely', () => {
    // 248..255 are above the largest multiple of the 31-letter alphabet and must be skipped.
    const bytes = [255, 250, 248, 0, 30, 31, 247]
    const random = (buffer: Uint8Array) => {
      buffer.fill(0)
      buffer.set(bytes.slice(0, buffer.length))
      return buffer
    }
    expect(generatePassword(4, random)).toBe('a9a9')
  })
})
