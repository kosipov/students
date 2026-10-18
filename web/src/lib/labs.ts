// Date helpers for the labs page. The university lives in Moscow time (UTC+3 all year),
// so the day arithmetic is done on the noon-UTC instant: it is the same calendar day in MSK.

function parseDay(date: string): Date {
  return new Date(`${date}T12:00:00Z`)
}

function formatDay(day: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${day.getUTCFullYear()}-${pad(day.getUTCMonth() + 1)}-${pad(day.getUTCDate())}`
}

/** Today's date in the university's time zone, "2026-10-19". */
export function mskToday(): string {
  return formatDay(new Date(Date.now() + 3 * 60 * 60 * 1000))
}

/** Monday of the week containing the date. */
export function mondayOf(date: string): string {
  const day = parseDay(date)
  day.setUTCDate(day.getUTCDate() - ((day.getUTCDay() + 6) % 7))
  return formatDay(day)
}

export function addDays(date: string, days: number): string {
  const day = parseDay(date)
  day.setUTCDate(day.getUTCDate() + days)
  return formatDay(day)
}

/** The range the campus site can show: the current week and three weeks ahead. */
export function campusDateRange(): { min: string; max: string } {
  const monday = mondayOf(mskToday())
  return { min: monday, max: addDays(monday, 27) }
}
