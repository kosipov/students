/** Picks the Russian plural form: plural(2, ['задание', 'задания', 'заданий']) === 'задания'. */
export function plural(count: number, forms: readonly [one: string, few: string, many: string]): string {
  const n = Math.abs(count) % 100
  const lastDigit = n % 10
  if (n >= 11 && n <= 19) return forms[2]
  if (lastDigit === 1) return forms[0]
  if (lastDigit >= 2 && lastDigit <= 4) return forms[1]
  return forms[2]
}

export const TASK_FORMS = ['задание', 'задания', 'заданий'] as const
export const SUBJECT_FORMS = ['предмет', 'предмета', 'предметов'] as const
export const GROUP_FORMS = ['группа', 'группы', 'групп'] as const

export function countLabel(count: number, forms: readonly [string, string, string]): string {
  return `${count} ${plural(count, forms)}`
}

/** "01", "02", … for card kickers. */
export function ordinal(index: number): string {
  return String(index + 1).padStart(2, '0')
}

/** Lowercases and treats "ё" as "е", so a search finds words typed either way. */
export function normalizeForSearch(value: string): string {
  return value.trim().toLocaleLowerCase('ru').replaceAll('ё', 'е')
}

const dateTimeFormat = new Intl.DateTimeFormat('ru-RU', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
})

const dateFormat = new Intl.DateTimeFormat('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' })

export function formatDateTime(iso: string): string {
  return dateTimeFormat.format(new Date(iso))
}

export function formatDate(iso: string): string {
  return dateFormat.format(new Date(iso))
}

/**
 * Keeps only paths inside the site, so a login link can't send the user to another site
 * after signing in (e.g. ?next=//evil.example).
 */
export function safeNextPath(next: string | null, fallback: string): string {
  if (!next || !next.startsWith('/') || next.startsWith('//') || next.startsWith('/\\')) {
    return fallback
  }
  return next
}
