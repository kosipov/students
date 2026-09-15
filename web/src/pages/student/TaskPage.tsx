import { useEffect, useState, type FormEvent } from 'react'
import { Link, useParams } from 'react-router'
import { ApiError } from '../../api/client'
import { usePreviewTask, useTask, useUnlockTask } from '../../api/queries'
import type { Task } from '../../api/types'
import { Breadcrumbs } from '../../components/Breadcrumbs'
import { Corners } from '../../components/Corners'
import { LockIcon } from '../../components/Icons'
import { Loading, QueryError } from '../../components/PageState'
import { TextField } from '../../components/TextField'
import { useSelection } from '../../layouts/StudentLayout'
import { useDocumentTitle } from '../../lib/hooks'
import { formatDate } from '../../lib/text'

const taskNotFound = { title: 'Задание не найдено', message: 'Возможно, его удалили или скрыли.' }

export function TaskPage() {
  const id = Number(useParams().taskId)
  const task = useTask(id)

  // A task that isn't a document is just a link: open it as students used to (after the password, if any).
  const href = task.data && !task.data.locked && !task.data.isDocument ? task.data.href : ''
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

function UnlockForm({ task }: { task: Task }) {
  const unlock = useUnlockTask(task.id)
  const [password, setPassword] = useState('')

  const submit = (event: FormEvent) => {
    event.preventDefault()
    unlock.mutate(password)
  }

  const error = unlock.error instanceof ApiError ? unlock.error.message : unlock.error ? 'Не удалось проверить пароль, попробуйте ещё раз' : ''

  return (
    <form className="blueprint unlock-card ri delay-2" onSubmit={submit} noValidate>
      <Corners />
      <div className="unlock-head">
        <LockIcon />
        <h2>Задание защищено паролем</h2>
      </div>
      <p className="muted">Введите пароль, который назвал преподаватель.</p>
      <TextField
        label="Пароль"
        type="password"
        value={password}
        onChange={setPassword}
        error={error}
        autoComplete="off"
        autoFocus
        maxLength={100}
      />
      <button type="submit" className="btn btn-primary" disabled={unlock.isPending || !password.trim()}>
        {unlock.isPending ? 'Проверяем…' : 'Открыть задание'}
      </button>
    </form>
  )
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
          {task.protected && ' Студенты откроют задание только после ввода пароля.'}
        </div>
      )}
      <h1 className="page-title ri">{task.name}</h1>
      {task.comment && <p className="subtitle ri delay-1">{task.comment}</p>}

      {task.locked ? <UnlockForm task={task} /> : <TaskBody task={task} subjectPath={subjectPath} />}
    </article>
  )
}

function TaskBody({ task, subjectPath }: { task: Task; subjectPath: string }) {
  return (
    <>
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
    </>
  )
}
