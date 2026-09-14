import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, ApiError } from './client'
import type {
  AdminGroup,
  AdminGroupDetails,
  AdminGroupsResponse,
  AdminSubject,
  AdminSubjectDetails,
  AdminTask,
  Catalog,
  GroupInput,
  Overview,
  SubjectInput,
  Task,
  TaskInput,
  User,
} from './types'

export const queryKeys = {
  catalog: ['catalog'] as const,
  task: (id: number) => ['task', id] as const,
  me: ['me'] as const,
  admin: ['admin'] as const,
  overview: ['admin', 'overview'] as const,
  groups: ['admin', 'groups'] as const,
  group: (id: number) => ['admin', 'group', id] as const,
  subject: (id: number) => ['admin', 'subject', id] as const,
  preview: (id: number) => ['admin', 'preview', id] as const,
  categories: ['admin', 'categories'] as const,
}

// Student

export function useCatalog() {
  return useQuery({ queryKey: queryKeys.catalog, queryFn: () => api.get<Catalog>('/catalog') })
}

export function useTask(id: number) {
  return useQuery({ queryKey: queryKeys.task(id), queryFn: () => api.get<Task>(`/tasks/${id}`) })
}

// Auth

/** The signed in user, or null for a guest. */
export function useMe() {
  return useQuery({
    queryKey: queryKeys.me,
    queryFn: async () => {
      try {
        return await api.get<User>('/auth/me')
      } catch (error) {
        if (error instanceof ApiError && error.status === 401) {
          return null
        }
        throw error
      }
    },
    staleTime: 5 * 60 * 1000,
  })
}

export function useSignIn() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (credentials: { login: string; password: string }) => api.post<User>('/auth/sign-in', credentials),
    onSuccess: (user) => queryClient.setQueryData(queryKeys.me, user),
  })
}

export function useSignOut() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<void>('/auth/sign-out'),
    onSuccess: () => {
      queryClient.setQueryData(queryKeys.me, null)
      queryClient.removeQueries({ queryKey: queryKeys.admin })
    },
  })
}

// Admin

export function useOverview() {
  return useQuery({ queryKey: queryKeys.overview, queryFn: () => api.get<Overview>('/admin/overview') })
}

export function useAdminGroups() {
  return useQuery({
    queryKey: queryKeys.groups,
    queryFn: async () => (await api.get<AdminGroupsResponse>('/admin/groups')).groups,
  })
}

export function useAdminGroup(id: number) {
  return useQuery({ queryKey: queryKeys.group(id), queryFn: () => api.get<AdminGroupDetails>(`/admin/groups/${id}`) })
}

export function useAdminSubject(id: number) {
  return useQuery({ queryKey: queryKeys.subject(id), queryFn: () => api.get<AdminSubjectDetails>(`/admin/subjects/${id}`) })
}

export function usePreviewTask(id: number) {
  return useQuery({ queryKey: queryKeys.preview(id), queryFn: () => api.get<Task>(`/admin/tasks/${id}/preview`) })
}

/** Category names already in use, suggested in the task form. */
export function useCategories() {
  return useQuery({
    queryKey: queryKeys.categories,
    queryFn: async () => (await api.get<{ categories: string[] }>('/admin/categories')).categories,
  })
}

/** Any change in the admin panel may affect every admin screen and what students see. */
function useInvalidateAll() {
  const queryClient = useQueryClient()
  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.admin }),
      queryClient.invalidateQueries({ queryKey: queryKeys.catalog }),
      queryClient.invalidateQueries({ queryKey: ['task'] }),
    ])
}

export function useCreateGroup() {
  const invalidate = useInvalidateAll()
  return useMutation({
    mutationFn: (input: GroupInput) => api.post<AdminGroup>('/admin/groups', input),
    onSuccess: invalidate,
  })
}

export function useUpdateGroup() {
  const invalidate = useInvalidateAll()
  return useMutation({
    mutationFn: ({ id, ...input }: GroupInput & { id: number }) => api.patch<AdminGroup>(`/admin/groups/${id}`, input),
    onSuccess: invalidate,
  })
}

export function useDeleteGroup() {
  const invalidate = useInvalidateAll()
  return useMutation({ mutationFn: (id: number) => api.delete(`/admin/groups/${id}`), onSuccess: invalidate })
}

export function useCreateSubject() {
  const invalidate = useInvalidateAll()
  return useMutation({
    mutationFn: ({ groupId, ...input }: SubjectInput & { groupId: number }) =>
      api.post<AdminSubject>(`/admin/groups/${groupId}/subjects`, input),
    onSuccess: invalidate,
  })
}

export function useUpdateSubject() {
  const invalidate = useInvalidateAll()
  return useMutation({
    mutationFn: ({ id, ...input }: SubjectInput & { id: number }) => api.patch<AdminSubject>(`/admin/subjects/${id}`, input),
    onSuccess: invalidate,
  })
}

export function useDeleteSubject() {
  const invalidate = useInvalidateAll()
  return useMutation({ mutationFn: (id: number) => api.delete(`/admin/subjects/${id}`), onSuccess: invalidate })
}

export function useCreateTask() {
  const invalidate = useInvalidateAll()
  return useMutation({
    mutationFn: ({ subjectId, ...input }: TaskInput & { subjectId: number }) =>
      api.post<AdminTask>(`/admin/subjects/${subjectId}/tasks`, input),
    onSuccess: invalidate,
  })
}

export function useUpdateTask() {
  const invalidate = useInvalidateAll()
  return useMutation({
    mutationFn: ({ id, ...input }: TaskInput & { id: number }) => api.patch<AdminTask>(`/admin/tasks/${id}`, input),
    onSuccess: invalidate,
  })
}

export function useDeleteTask() {
  const invalidate = useInvalidateAll()
  return useMutation({ mutationFn: (id: number) => api.delete(`/admin/tasks/${id}`), onSuccess: invalidate })
}

export function useRefreshTask() {
  const invalidate = useInvalidateAll()
  return useMutation({
    mutationFn: (id: number) => api.post<AdminTask>(`/admin/tasks/${id}/refresh`),
    onSuccess: invalidate,
  })
}
