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
