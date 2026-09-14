import { useEffect } from 'react'
import { Link, useParams } from 'react-router'
import { usePreviewTask, useTask } from '../../api/queries'
import type { Task } from '../../api/types'
import { Breadcrumbs } from '../../components/Breadcrumbs'
import { Loading, QueryError } from '../../components/PageState'
import { useSelection } from '../../layouts/StudentLayout'
import { useDocumentTitle } from '../../lib/hooks'
import { formatDate } from '../../lib/text'

const taskNotFound = { title: 'Задание не найдено', message: 'Возможно, его удалили или скрыли.' }

export function TaskPage() {
  const id = Number(useParams().taskId)
  const task = useTask(id)

  // A task that isn't a document is just a link: open it as students used to.
  const href = task.data && !task.data.isDocument ? task.data.href : ''
  useEffect(() => {
    if (href) window.location.replace(href)
  }, [href])

  if (task.isPending) return <Loading />
  if (task.isError) return <QueryError error={task.error} retry={() => task.refetch()} notFound={taskNotFound} />
  if (href) return <p className="loading">Открываем ссылку…</p>
  return <TaskView task={task.data} />
}

/** The task page for the admin: shows hidden tasks and doesn't redirect links. */
export function TaskPreviewPage() {
  const id = Number(useParams().taskId)
  const task = usePreviewTask(id)

  if (task.isPending) return <Loading />
  if (task.isError) {
    return (
      <QueryError
        error={task.error}
        retry={() => task.refetch()}
        notFound={{ ...taskNotFound, backTo: '/admin', backLabel: '← В управление' }}
      />
    )
  }
  return <TaskView task={task.data} preview />
}

function TaskView({ task, preview = false }: { task: Task; preview?: boolean }) {
  useSelection(preview ? undefined : task.group.id, preview ? undefined : task.subject.id)
  useDocumentTitle(task.name)

  const subjectPath = preview ? `/admin/subjects/${task.subject.id}` : `/groups/${task.group.id}/subjects/${task.subject.id}`

  return (
    <article className="task-page">
      <Breadcrumbs
        items={[
          { label: 'Группы', to: preview ? '/admin/groups' : '/' },
          { label: task.subject.name, to: subjectPath },
          { label: task.name },
        ]}
      />
      {preview && (
        <div className="notice fi">
          Предпросмотр для администратора.{task.hidden && ' Задание скрыто — студенты его не видят.'}
        </div>
      )}
      <h1 className="page-title ri">{task.name}</h1>
      {task.comment && <p className="subtitle ri delay-1">{task.comment}</p>}

      <div className="actions-row ri delay-1">
        {task.href && (
          <a className="btn btn-primary" href={task.href} target="_blank" rel="noopener noreferrer">
            Открыть файл
          </a>
        )}
        <Link className="btn btn-secondary" to={subjectPath}>
          ← Все задания
        </Link>
      </div>

      {!task.isDocument ? (
        <p className="muted">Это задание открывается по ссылке.</p>
      ) : task.contentHtml === null ? (
        <div className="notice notice--warning" role="alert">
          Не удалось загрузить задание. Попробуйте позже или откройте файл по кнопке выше.
        </div>
      ) : (
        <>
          {/* The server renders markdown without raw HTML and dangerous links, so it is safe to insert. */}
          <div className="md ri delay-2" dangerouslySetInnerHTML={{ __html: task.contentHtml }} />
          {task.updatedAt && <p className="muted">Обновлено {formatDate(task.updatedAt)}</p>}
        </>
      )}
    </article>
  )
}
