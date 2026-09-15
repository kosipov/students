import { describe, expect, it } from 'vitest'
import type { AdminTask } from '../../api/types'
import { taskStatus } from './taskStatus'

const base: AdminTask = {
  id: 1,
  subjectId: 1,
  name: 'Задание',
  href: 'https://1drv.ms/t/c/1/abc',
  comment: '',
  hidden: false,
  categories: [],
  password: '',
  isDocument: true,
  contentUnsupported: false,
  fetchedAt: null,
  checkedAt: null,
  error: '',
}

describe('taskStatus', () => {
  it('reports a downloaded document', () => {
    expect(taskStatus({ ...base, fetchedAt: '2026-09-12T09:14:00Z' })).toMatchObject({ isError: false, text: expect.stringMatching(/^Загружено 12\.09\.2026/) })
  })

  it('says students see the previous version when refreshing fails', () => {
    const status = taskStatus({ ...base, fetchedAt: '2026-09-12T09:14:00Z', error: 'onedrive is down' })
    expect(status.isError).toBe(true)
    expect(status.text).toContain('студенты видят версию от')
  })

  it('says students see nothing when the first download fails', () => {
    expect(taskStatus({ ...base, error: 'onedrive is down' })).toEqual({
      text: 'Не удалось загрузить файл — студенты его пока не видят',
      isError: true,
    })
  })

  it('describes links that are not documents', () => {
    expect(taskStatus({ ...base, isDocument: false, contentUnsupported: true }).text).toContain('Не markdown-файл')
    expect(taskStatus({ ...base, isDocument: false, href: 'https://example.com' }).text).toContain('Внешняя ссылка')
    expect(taskStatus({ ...base, isDocument: false, href: '' }).text).toContain('Добавьте ссылку')
  })

  it('reports a document that has not been downloaded yet', () => {
    expect(taskStatus(base).text).toBe('Ещё не загружалось')
  })
})
