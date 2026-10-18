import { Navigate, Route, Routes } from 'react-router'
import { AdminLayout, RequireAdmin } from './layouts/AdminLayout'
import { StudentLayout } from './layouts/StudentLayout'
import { GroupPage } from './pages/admin/GroupPage'
import { GroupsPage } from './pages/admin/GroupsPage'
import { LabsPage } from './pages/admin/LabsPage'
import { OverviewPage } from './pages/admin/OverviewPage'
import { ProgressPage } from './pages/admin/ProgressPage'
import { SubjectPage as AdminSubjectPage } from './pages/admin/SubjectPage'
import { LoginPage } from './pages/LoginPage'
import { NotFound } from './components/PageState'
import { HomePage } from './pages/student/HomePage'
import { ProgressGroupsPage, ProgressRosterPage, StudentProgressPage } from './pages/student/ProgressPages'
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
        <Route path="progress" element={<ProgressGroupsPage />} />
        <Route path="progress/groups/:groupId" element={<ProgressRosterPage />} />
        <Route path="progress/students/:studentId" element={<StudentProgressPage />} />
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
        <Route path="labs" element={<LabsPage />} />
        <Route path="groups" element={<GroupsPage />} />
        <Route path="groups/:groupId" element={<GroupPage />} />
        <Route path="groups/:groupId/progress" element={<ProgressPage />} />
        <Route path="subjects/:subjectId" element={<AdminSubjectPage />} />
        <Route path="*" element={<NotFound backTo="/admin" backLabel="← В управление" />} />
      </Route>
    </Routes>
  )
}
