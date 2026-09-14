import type { Catalog, CatalogGroup, CatalogSubject, CatalogTask } from '../api/types'
import { normalizeForSearch } from './text'

export function findGroup(catalog: Catalog, groupId: number): CatalogGroup | undefined {
  return catalog.groups.find((group) => group.id === groupId)
}

export function findSubject(group: CatalogGroup, subjectId: number): CatalogSubject | undefined {
  return group.subjects.find((subject) => subject.id === subjectId)
}

export function countTasks(group: CatalogGroup): number {
  return group.subjects.reduce((count, subject) => count + subject.tasks.length, 0)
}

export interface SearchResult {
  task: CatalogTask
  group: CatalogGroup
  subject: CatalogSubject
}

/** Finds tasks whose name or subject name contains the query. */
export function searchTasks(catalog: Catalog, query: string): SearchResult[] {
  const needle = normalizeForSearch(query)
  if (!needle) return []

  const results: SearchResult[] = []
  for (const group of catalog.groups) {
    for (const subject of group.subjects) {
      const subjectMatches = normalizeForSearch(subject.name).includes(needle)
      for (const task of subject.tasks) {
        if (subjectMatches || normalizeForSearch(task.name).includes(needle)) {
          results.push({ task, group, subject })
        }
      }
    }
  }
  return results
}

/** Treats "Курсовые" and "курсовые" (and "ё"/"е") as one category, like the server does. */
export function categoryKey(name: string): string {
  return normalizeForSearch(name)
}

/** Categories used by the tasks, each once, sorted alphabetically. */
export function categoriesOf(tasks: CatalogTask[]): string[] {
  const byKey = new Map<string, string>()
  for (const task of tasks) {
    for (const category of task.categories) {
      const key = categoryKey(category)
      if (!byKey.has(key)) byKey.set(key, category)
    }
  }
  return Array.from(byKey.values()).sort((a, b) => a.localeCompare(b, 'ru'))
}

/** Tasks with the category; all tasks when no category is chosen. */
export function filterByCategory(tasks: CatalogTask[], category: string | null): CatalogTask[] {
  if (!category) return tasks
  const key = categoryKey(category)
  return tasks.filter((task) => task.categories.some((c) => categoryKey(c) === key))
}
