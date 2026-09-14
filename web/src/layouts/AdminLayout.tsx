import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import { Link, Navigate, Outlet, useLocation } from 'react-router'
import { useMe } from '../api/queries'
import { Footer, Header } from '../components/Header'
import { Loading } from '../components/PageState'

interface AdminContext {
  groupId?: number
  subjectId?: number
}

interface AdminShell {
  /** Pages report the group and subject they show, so "Предметы" and "Задания" lead back to them. */
  remember: (context: AdminContext) => void
}

const AdminShellContext = createContext<AdminShell | null>(null)

export function useRememberAdminContext(groupId: number | undefined, subjectId?: number) {
  const shell = useContext(AdminShellContext)
  useEffect(() => {
    if (groupId !== undefined) {
      shell?.remember({ groupId, subjectId })
    }
  }, [shell, groupId, subjectId])
}

type Section = 'overview' | 'groups' | 'subjects' | 'tasks'

function currentSection(pathname: string): Section {
  if (pathname.startsWith('/admin/subjects')) return 'tasks'
  if (/^\/admin\/groups\/\d+/.test(pathname)) return 'subjects'
  if (pathname.startsWith('/admin/groups')) return 'groups'
  return 'overview'
}

/** Lets only the signed in admin through; a guest is sent to the login page and back. */
export function RequireAdmin({ children }: { children: React.ReactNode }) {
  const me = useMe()
  const location = useLocation()

  if (me.isPending) {
    return <Loading />
  }
  if (!me.data) {
    const next = encodeURIComponent(location.pathname + location.search)
    return <Navigate to={`/login?next=${next}`} replace />
  }
  return <>{children}</>
}

export function AdminLayout() {
  const location = useLocation()
  const [context, setContext] = useState<AdminContext>({})
  const shell = useMemo<AdminShell>(
    () => ({
      remember: (next) =>
        setContext((prev) => ({
          groupId: next.groupId,
          // On a group page keep the subject opened before, unless it belongs to another group.
          subjectId: next.subjectId ?? (next.groupId === prev.groupId ? prev.subjectId : undefined),
        })),
    }),
    [],
  )

  const section = currentSection(location.pathname)
  const nav: { key: Section; label: string; to: string }[] = [
    { key: 'overview', label: 'Обзор', to: '/admin' },
    { key: 'groups', label: 'Группы', to: '/admin/groups' },
    { key: 'subjects', label: 'Предметы', to: context.groupId ? `/admin/groups/${context.groupId}` : '/admin/groups' },
    { key: 'tasks', label: 'Задания', to: context.subjectId ? `/admin/subjects/${context.subjectId}` : context.groupId ? `/admin/groups/${context.groupId}` : '/admin/groups' },
  ]

  return (
    <RequireAdmin>
      <div className="shell">
        <Header isAdminArea />

        <nav className="mobnav" aria-label="Управление">
          {nav.map((item) => (
            <Link key={item.key} to={item.to} className={`btn ${item.key === section ? 'btn-primary' : 'btn-secondary'}`}>
              {item.label}
            </Link>
          ))}
          <Link to="/" className="btn btn-secondary">
            Как видят студенты
          </Link>
        </nav>

        <div className="layout">
          <aside className="sidebar">
            <nav aria-label="Управление">
              <div className="sidebar-label">Управление</div>
              {nav.map((item) => (
                <Link
                  key={item.key}
                  to={item.to}
                  className={`navitem si navitem-title ${item.key === section ? 'is-active' : ''}`}
                  aria-current={item.key === section ? 'page' : undefined}
                >
                  {item.label}
                </Link>
              ))}
              <div className="sidebar-divider" />
              <Link to="/" className="navitem navitem-sub">
                ← Как видят студенты
              </Link>
            </nav>
          </aside>

          <main className="main" id="main">
            <AdminShellContext.Provider value={shell}>
              <Outlet />
            </AdminShellContext.Provider>
          </main>
        </div>

        <Footer />
      </div>
    </RequireAdmin>
  )
}
