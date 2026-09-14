import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { ApiError } from '../../api/client'
import { useAdminSubject, useDeleteTask, useRefreshTask, useUpdateTask } from '../../api/queries'
import type { AdminTask } from '../../api/types'
import { Breadcrumbs } from '../../components/Breadcrumbs'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { Corners } from '../../components/Corners'
import { PlusIcon } from '../../components/Icons'
import { Loading, QueryError } from '../../components/PageState'
import { useToast } from '../../components/Toast'
import { useRememberAdminContext } from '../../layouts/AdminLayout'
import { useDocumentTitle } from '../../lib/hooks'
import { taskStatus } from './taskStatus'
import { TaskDialog } from './dialogs'

type Dialog = { kind: 'edit'; task?: AdminTask } | { kind: 'delete'; task: AdminTask }

const subjectNotFound = { title: 'Предмет не найден', backTo: '/admin/groups', backLabel: '← К группам' }

export function SubjectPage() {
  const id = Number(useParams().subjectId)
  const details = useAdminSubject(id)
  const update = useUpdateTask()
  const refresh = useRefreshTask()
  const remove = useDeleteTask()
  const toast = useToast()
  const [dialog, setDialog] = useState<Dialog | null>(null)

  useRememberAdminContext(details.data?.group.id, details.data?.subject.id)
  useDocumentTitle(details.data ? `Задания: ${details.data.subject.name}` : undefined)

  if (details.isPending) return <Loading />
  if (details.isError) return <QueryError error={details.error} retry={() => details.refetch()} notFound={subjectNotFound} />

  const { group, subject, tasks } = details.data

  const toggle = (task: AdminTask) =>
    update.mutate(
      { id: task.id, hidden: !task.hidden },
      {
        onSuccess: (saved) => toast(saved.hidden ? 'Задание скрыто' : 'Задание видно студентам'),
        onError: () => toast('Не удалось изменить видимость'),
      },
    )

  const refreshTask = (task: AdminTask) =>
    refresh.mutate(task.id, {
      onSuccess: (saved) =>
        toast(saved.contentUnsupported ? 'Это не markdown-файл, он открывается по ссылке' : saved.error ? 'Файл по-прежнему недоступен' : 'Содержимое обновлено'),
      onError: (error) => toast(error instanceof ApiError ? error.message : 'Не удалось обновить'),
    })

  return (
    <div>
      <Breadcrumbs
        items={[
          { label: 'Группы', to: '/admin/groups' },
          { label: group.name, to: `/admin/groups/${group.id}` },
          { label: subject.name },
        ]}
      />
      <div className="page-head">
        <div className="title-with-tag">
          <h1 className="page-title">Задания</h1>
          {group.hidden && <span className="tag tag-neutral">Группа скрыта</span>}
        </div>
        <button type="button" className="btn btn-primary" onClick={() => setDialog({ kind: 'edit' })}>
          <PlusIcon />
          Задание
        </button>
      </div>

      {tasks.length === 0 ? (
        <p className="muted">В предмете пока нет заданий.</p>
      ) : (
        <div className="task-rows stg">
          {tasks.map((task) => {
            const status = taskStatus(task)
            const canRefresh = task.isDocument || task.contentUnsupported
            return (
              <div key={task.id} className="blueprint task-row">
                <Corners />
                <div>
                  <div className="task-row-head">
                    <span className="task-row-name">{task.name}</span>
                    <span className={`tag ${task.hidden ? 'tag-neutral' : 'tag-accent'}`}>{task.hidden ? 'Скрыто' : 'Видно'}</span>
                  </div>
                  <div className="task-row-href">{task.href || 'ссылка не указана'}</div>
                  {task.comment && <div className="task-row-comment">{task.comment}</div>}
                  <div className={`task-row-status ${status.isError ? 'is-error' : ''}`} title={task.error || undefined}>
                    {status.text}
                  </div>
                </div>
                <div className="task-row-actions">
                  {task.isDocument ? (
                    <Link className="btn btn-secondary" to={`/admin/tasks/${task.id}/preview`}>
                      Открыть
                    </Link>
                  ) : (
                    task.href && (
                      <a className="btn btn-secondary" href={task.href} target="_blank" rel="noopener noreferrer">
                        Открыть
                      </a>
                    )
                  )}
                  {canRefresh && (
                    <button
                      type="button"
                      className="btn btn-secondary"
                      onClick={() => refreshTask(task)}
                      disabled={refresh.isPending && refresh.variables === task.id}
                    >
                      {refresh.isPending && refresh.variables === task.id ? 'Обновляем…' : 'Обновить'}
                    </button>
                  )}
                  <button type="button" className="btn btn-secondary" onClick={() => toggle(task)}>
                    {task.hidden ? 'Показать' : 'Скрыть'}
                  </button>
                  <button type="button" className="btn btn-primary" onClick={() => setDialog({ kind: 'edit', task })}>
                    Изменить
                  </button>
                  <button type="button" className="btn btn-ghost" onClick={() => setDialog({ kind: 'delete', task })}>
                    Удалить
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {dialog?.kind === 'edit' && <TaskDialog subjectId={subject.id} task={dialog.task} onClose={() => setDialog(null)} />}
      {dialog?.kind === 'delete' && (
        <ConfirmDialog
          title={`Удалить задание «${dialog.task.name}»?`}
          message="Это нельзя отменить."
          confirmLabel="Удалить"
          onConfirm={() => remove.mutateAsync(dialog.task.id).then(() => toast('Задание удалено'))}
          onClose={() => setDialog(null)}
        />
      )}
    </div>
  )
}
