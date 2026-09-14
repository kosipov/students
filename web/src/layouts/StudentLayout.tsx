import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import { Link, Outlet, useLocation } from 'react-router'
import { useCatalog } from '../api/queries'
import type { Catalog } from '../api/types'
import { Footer, Header } from '../components/Header'
import { Loading, QueryError } from '../components/PageState'
import { countTasks, findGroup } from '../lib/catalog'
import { SearchResults } from '../pages/student/SearchResults'

interface Selection {
  groupId?: number
  subjectId?: number
}

interface StudentShell {
  catalog: Catalog
  /** Pages report the group and subject they show, so the sidebar can highlight them. */
  select: (selection: Selection) => void
}

const StudentShellContext = createContext<StudentShell | null>(null)

export function useStudentShell(): StudentShell {
  const shell = useContext(StudentShellContext)
  if (!shell) {
    throw new Error('useStudentShell must be used inside StudentLayout')
  }
  return shell
}

/** Highlights the group and subject in the sidebar while the page is shown. */
export function useSelection(groupId: number | undefined, subjectId: number | undefined) {
  const { select } = useStudentShell()
  useEffect(() => {
    select({ groupId, subjectId })
  }, [select, groupId, subjectId])
}

export function StudentLayout() {
  const catalog = useCatalog()
  const location = useLocation()
  const [query, setQuery] = useState('')
  const [selection, setSelection] = useState<Selection>({})

  // A new page starts without the previous search.
  useEffect(() => setQuery(''), [location.pathname])

  const shell = useMemo<StudentShell | null>(
    () => (catalog.data ? { catalog: catalog.data, select: setSelection } : null),
    [catalog.data],
  )

  const groups = catalog.data?.groups ?? []
  const activeGroup = catalog.data && selection.groupId !== undefined ? findGroup(catalog.data, selection.groupId) : undefined

  return (
    <div className="shell">
      <Header search={{ value: query, onChange: setQuery }} />

      <nav className="mobnav" aria-label="Группы">
        {groups.map((group) => (
          <Link
            key={group.id}
            to={`/groups/${group.id}`}
            className={`btn ${group.id === activeGroup?.id ? 'btn-primary' : 'btn-secondary'}`}
            aria-current={group.id === activeGroup?.id ? 'true' : undefined}
          >
            {group.name}
          </Link>
        ))}
      </nav>

      <div className="layout">
        <aside className="sidebar">
          <nav aria-label="Группы">
            <div className="sidebar-label">Группы</div>
            {groups.map((group) => (
              <Link
                key={group.id}
                to={`/groups/${group.id}`}
                className={`navitem si ${group.id === activeGroup?.id ? 'is-active' : ''}`}
                aria-current={group.id === activeGroup?.id ? 'true' : undefined}
              >
                <span className="navitem-title">{group.name}</span>
                <span className="tag tag-neutral" title="Заданий">
                  {countTasks(group)}
                </span>
              </Link>
            ))}
          </nav>

          {activeGroup && activeGroup.subjects.length > 0 && (
            <nav aria-label="Предметы">
              <div className="sidebar-divider" />
              <div className="sidebar-label">Предметы</div>
              {activeGroup.subjects.map((subject) => (
                <Link
                  key={subject.id}
                  to={`/groups/${activeGroup.id}/subjects/${subject.id}`}
                  className={`navitem navitem-sub ${subject.id === selection.subjectId ? 'is-active' : ''}`}
                  aria-current={subject.id === selection.subjectId ? 'true' : undefined}
                >
                  {subject.name}
                </Link>
              ))}
            </nav>
          )}
        </aside>

        <main className="main" id="main">
          {catalog.isPending && <Loading />}
          {catalog.isError && <QueryError error={catalog.error} retry={() => catalog.refetch()} />}
          {shell && (
            <StudentShellContext.Provider value={shell}>
              {query.trim() ? <SearchResults catalog={shell.catalog} query={query} /> : <Outlet />}
            </StudentShellContext.Provider>
          )}
        </main>
      </div>

      <Footer />
    </div>
  )
}
