import { Link } from 'react-router'
import { ApiError } from '../../api/client'
import { useOverview, useRefreshTask } from '../../api/queries'
import { Corners } from '../../components/Corners'
import { Loading, QueryError } from '../../components/PageState'
import { useToast } from '../../components/Toast'
import { useCountUp, useDocumentTitle } from '../../lib/hooks'
import { formatDate } from '../../lib/text'

function Stat({ value, label, index }: { value: number; label: string; index: number }) {
  const shown = useCountUp(value, index * 80)
  return (
    <div className="blueprint stat">
      <Corners />
      <div className="stat-value" aria-hidden="true">
        {shown}
      </div>
      <div className="stat-label">
        <span className="visually-hidden">{value} </span>
        {label}
      </div>
    </div>
  )
}

export function OverviewPage() {
  useDocumentTitle('Обзор')
  const overview = useOverview()
  const refresh = useRefreshTask()
  const toast = useToast()

  if (overview.isPending) return <Loading />
  if (overview.isError) return <QueryError error={overview.error} retry={() => overview.refetch()} />

  const { stats, recentlyUpdated, issues } = overview.data
  const statItems = [
    { label: 'групп', value: stats.groups },
    { label: 'предметов', value: stats.subjects },
    { label: 'заданий', value: stats.tasks },
    { label: 'скрыто', value: stats.hidden },
  ]

  const refreshTask = (id: number) =>
    refresh.mutate(id, {
      onSuccess: (task) => toast(task.error ? 'Файл по-прежнему недоступен' : 'Содержимое обновлено'),
      onError: (error) => toast(error instanceof ApiError ? error.message : 'Не удалось обновить'),
    })

  return (
    <div>
      <h1 className="page-title ri">Обзор</h1>

      <div className="stats stg">
        {statItems.map((item, index) => (
          <Stat key={item.label} index={index} {...item} />
        ))}
      </div>

      <div className="grid-2">
        <section className="blueprint panel ri" aria-labelledby="recent-title">
          <Corners />
          <h3 id="recent-title">Недавно обновлённые файлы</h3>
          {recentlyUpdated.length === 0 ? (
            <p className="muted">Файлы из OneDrive ещё не загружались.</p>
          ) : (
            <div className="stg">
              {recentlyUpdated.map((task) => (
                <Link key={task.id} to={`/admin/subjects/${task.subjectId}`} className="panel-row">
                  <span>
                    <span className="panel-row-title">{task.name}</span>
                    <span className="panel-row-sub">
                      {task.groupName} · {task.subjectName}
                    </span>
                  </span>
                  {task.fetchedAt && <span className="muted">{formatDate(task.fetchedAt)}</span>}
                </Link>
              ))}
            </div>
          )}
        </section>

        <section className="blueprint panel ri delay-1" aria-labelledby="issues-title">
          <Corners />
          <h3 id="issues-title">Требует внимания</h3>
          {issues.length === 0 ? (
            <p className="muted">Всё в порядке.</p>
          ) : (
            issues.map((task) => (
              <div key={task.id} className="panel-row">
                <span>
                  <Link to={`/admin/subjects/${task.subjectId}`} className="panel-row-title">
                    {task.name}
                  </Link>
                  <span className="panel-row-problem">
                    Не удалось загрузить файл{task.fetchedAt ? ' — студенты видят прошлую версию' : ''}
                  </span>
                </span>
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => refreshTask(task.id)}
                  disabled={refresh.isPending && refresh.variables === task.id}
                >
                  Обновить
                </button>
              </div>
            ))
          )}
        </section>
      </div>
    </div>
  )
}
