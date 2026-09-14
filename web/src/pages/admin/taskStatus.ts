import type { AdminTask } from '../../api/types'
import { formatDateTime } from '../../lib/text'

export interface TaskStatus {
  text: string
  isError: boolean
}

/** Describes what students get for the task and whether its file is fine. */
export function taskStatus(task: AdminTask): TaskStatus {
  if (!task.href) {
    return { text: 'Добавьте ссылку, чтобы студенты могли открыть задание', isError: false }
  }
  if (task.contentUnsupported) {
    return { text: 'Не markdown-файл — студенты открывают его по ссылке', isError: false }
  }
  if (!task.isDocument) {
    return { text: 'Внешняя ссылка — открывается как есть', isError: false }
  }
  if (task.error) {
    return {
      text: task.fetchedAt
        ? `Не удалось обновить файл — студенты видят версию от ${formatDateTime(task.fetchedAt)}`
        : 'Не удалось загрузить файл — студенты его пока не видят',
      isError: true,
    }
  }
  if (task.fetchedAt) {
    return { text: `Загружено ${formatDateTime(task.fetchedAt)}`, isError: false }
  }
  return { text: 'Ещё не загружалось', isError: false }
}
