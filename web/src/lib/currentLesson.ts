import type { Catalog, CatalogGroup, CatalogSubject, ScheduleLesson } from '../api/types'
import { normalizeForSearch } from './text'

export interface LessonMatch {
  group: CatalogGroup
  subject: CatalogSubject
}

/** Group names of a lesson: the site writes them one per line or comma separated, e.g. "1435-ИРо, 1425-ИРо". */
function lessonGroups(lesson: ScheduleLesson): string[] {
  return lesson.groups
    .split(/[,;\n]/)
    .map((name) => normalizeForSearch(name))
    .filter(Boolean)
}

/**
 * Finds the group and subject of a lesson among what students can see. Names come from the university's
 * site and from the admin panel, so they are compared ignoring case, "ё" and extra spaces.
 */
export function matchLesson(catalog: Catalog, lesson: ScheduleLesson | null): LessonMatch | null {
  if (!lesson) return null

  const groups = lessonGroups(lesson)
  const title = normalizeForSearch(lesson.title)
  if (!title) return null

  const candidates = groups.length > 0 ? catalog.groups.filter((group) => groups.includes(normalizeForSearch(group.name))) : catalog.groups

  for (const group of candidates) {
    const subject = group.subjects.find((s) => normalizeForSearch(s.name) === title)
    if (subject) return { group, subject }
  }
  // The subject may be named a little differently, e.g. with a shortened word.
  for (const group of candidates) {
    const subject = group.subjects.find((s) => {
      const name = normalizeForSearch(s.name)
      return name.startsWith(title) || title.startsWith(name)
    })
    if (subject) return { group, subject }
  }
  return null
}

/** Puts the group that has a lesson right now first; the rest keep their order. */
export function groupsWithLessonFirst(groups: CatalogGroup[], match: LessonMatch | null): CatalogGroup[] {
  if (!match) return groups
  const current = groups.find((group) => group.id === match.group.id)
  if (!current) return groups
  return [current, ...groups.filter((group) => group.id !== current.id)]
}
