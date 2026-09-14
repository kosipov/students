import { Navigate, Route, Routes } from 'react-router'
import { AdminLayout, RequireAdmin } from './layouts/AdminLayout'
import { StudentLayout } from './layouts/StudentLayout'
import { GroupPage } from './pages/admin/GroupPage'
import { GroupsPage } from './pages/admin/GroupsPage'
import { OverviewPage } from './pages/admin/OverviewPage'
import { SubjectPage as AdminSubjectPage } from './pages/admin/SubjectPage'
import { LoginPage } from './pages/LoginPage'
import { NotFound } from './components/PageState'
import { HomePage } from './pages/student/HomePage'
import { SubjectPage } from './pages/student/SubjectPage'
import { TaskPage, TaskPreviewPage } from './pages/student/TaskPage'

export function App() {
  return (
    <Routes>
      <Route element={<StudentLayout />}>
        <Route index element={<HomePage />} />
        <Route path="groups/:groupId" element={<HomePage />} />
        <Route path="groups/:groupId/subjects/:subjectId" element={<SubjectPage />} />
        <Route path="tasks/:taskId" element={<TaskPage />} />
        <Route
          path="admin/tasks/:taskId/preview"
          element={
            <RequireAdmin>
              <TaskPreviewPage />
            </RequireAdmin>
          }
        />
        <Route path="*" element={<NotFound />} />
      </Route>

      <Route path="login" element={<LoginPage />} />
      {/* The address of the old server-rendered login page. */}
      <Route path="auth/sign-in" element={<Navigate to="/login" replace />} />

      <Route path="admin" element={<AdminLayout />}>
        <Route index element={<OverviewPage />} />
        <Route path="groups" element={<GroupsPage />} />
        <Route path="groups/:groupId" element={<GroupPage />} />
        <Route path="subjects/:subjectId" element={<AdminSubjectPage />} />
        <Route path="*" element={<NotFound backTo="/admin" backLabel="← В управление" />} />
      </Route>
    </Routes>
  )
}
