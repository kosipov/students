// Types of the Go API responses (educational/delivery/http/dto.go).

export interface Ref {
  id: number
  name: string
}

export interface CatalogTask {
  id: number
  name: string
  comment: string
  href: string
  /** Shown as a page on the site; other tasks are opened by href. */
  isDocument: boolean
  categories: string[]
  /** Protected by a password the student hasn't entered; href is empty. */
  locked: boolean
}

export interface CatalogSubject {
  id: number
  name: string
  tasks: CatalogTask[]
}

export interface CatalogGroup {
  id: number
  name: string
  subjects: CatalogSubject[]
}

export interface Catalog {
  groups: CatalogGroup[]
}

export interface Task {
  id: number
  name: string
  comment: string
  href: string
  hidden: boolean
  isDocument: boolean
  protected: boolean
  /** The password hasn't been entered: no href and no document. */
  locked: boolean
  group: Ref
  subject: Ref
  /** Rendered document, null when it has never been downloaded. */
  contentHtml: string | null
  updatedAt: string | null
}

export interface User {
  userName: string
}

export interface AdminGroup {
  id: number
  name: string
  hidden: boolean
  subjectsCount: number
  tasksCount: number
}

export interface AdminSubject {
  id: number
  groupId: number
  name: string
  tasksCount: number
}

export interface AdminTask {
  id: number
  subjectId: number
  name: string
  href: string
  comment: string
  hidden: boolean
  categories: string[]
  /** Empty when the task has no password. */
  password: string
  isDocument: boolean
  contentUnsupported: boolean
  fetchedAt: string | null
  checkedAt: string | null
  error: string
}

export interface AdminGroupsResponse {
  groups: AdminGroup[]
}

export interface AdminGroupDetails {
  group: AdminGroup
  subjects: AdminSubject[]
}

export interface AdminSubjectDetails {
  group: AdminGroup
  subject: AdminSubject
  tasks: AdminTask[]
}

export interface OverviewTask {
  id: number
  name: string
  groupName: string
  subjectId: number
  subjectName: string
  fetchedAt: string | null
  error: string
}

export interface Overview {
  stats: { groups: number; subjects: number; tasks: number; hidden: number }
  recentlyUpdated: OverviewTask[]
  issues: OverviewTask[]
}

export interface GroupInput {
  name?: string
  hidden?: boolean
}

export interface SubjectInput {
  name: string
}

export interface TaskInput {
  name?: string
  href?: string
  comment?: string
  hidden?: boolean
  /** Replaces all categories of the task. */
  categories?: string[]
  /** An empty password removes the protection. */
  password?: string
}
