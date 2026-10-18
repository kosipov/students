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

export interface ScheduleLesson {
  number: number
  title: string
  /** As on the university's site: "л.", "пр.", "лаб.". */
  kind: string
  location: string
  room: string
  building: string
  groups: string
  startsAt: string
  endsAt: string
}

export type PresenceStatus = 'in_class' | 'break' | 'away' | 'unknown'

export interface Presence {
  status: PresenceStatus
  current: ScheduleLesson | null
  /** The next lesson today. */
  next: ScheduleLesson | null
  syncedAt: string | null
  /** The schedule hasn't been refreshed for a long time and may have changed. */
  stale: boolean
}

export interface ScheduleSyncState {
  checkedAt: string | null
  succeededAt: string | null
  campusUpdatedAt: string | null
  error: string
  stale: boolean
}

// Progress: the grade sheet (progress/delivery/http/dto.go).

export type MarkStatus = 'in_progress' | 'submitted' | 'accepted'

export interface ProgressMark {
  taskId: number
  status: MarkStatus
  comment?: string
  updatedAt: string
}

export interface ProgressStudent {
  id: number
  name: string
  marks: ProgressMark[]
}

export interface ProgressTask {
  id: number
  name: string
  hidden: boolean
}

export interface ProgressSubject {
  id: number
  name: string
  tasks: ProgressTask[]
}

/** The admin matrix: students as rows, tasks of each subject as columns. */
export interface AdminProgress {
  group: { id: number; name: string; hidden: boolean }
  subjects: ProgressSubject[]
  students: ProgressStudent[]
}

/** The public roster: names of the group's students, without marks. */
export interface Roster {
  group: Ref
  students: Ref[]
}

export interface StudentProgressTask {
  id: number
  name: string
  comment: string
  /** Empty when the teacher hasn't marked anything yet. */
  status: MarkStatus | ''
  markComment?: string
  updatedAt: string | null
}

export interface StudentProgressSubject {
  id: number
  name: string
  tasks: StudentProgressTask[]
}

/** What a student sees on the public progress page. */
export interface StudentProgress {
  student: Ref
  group: Ref
  subjects: StudentProgressSubject[]
}

export interface ImportResult {
  added: number
  skipped: number
  /** How many names the source had (the pasted list or the campus roster). */
  found: number
}

// Labs occupancy (admin only)

export type LabsFloor = '2' | '5' | 'all'

export interface LabLesson {
  number: number
  title: string
  kind: string
  teacher: string
  groups: string
  startsAt: string
  endsAt: string
}

export interface LabRoom {
  /** The full campus name with the building, e.g. "245/1". */
  room: string
  floor: number
  busy: boolean
  lesson: LabLesson | null
  fetchedAt: string | null
  /** The campus didn't answer; an older copy is shown. */
  stale: boolean
  /** The campus didn't answer and there is nothing to show. */
  failed: boolean
}

export interface LabsOccupancy {
  date: string
  pair: number
  startsAt: string
  endsAt: string
  /** The oldest campus answer among the rooms; null when nothing loaded at all. */
  fetchedAt: string | null
  rooms: LabRoom[]
}
