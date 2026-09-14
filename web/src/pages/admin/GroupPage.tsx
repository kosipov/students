import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router'
import { useAdminGroup, useDeleteSubject } from '../../api/queries'
import type { AdminSubject } from '../../api/types'
import { Breadcrumbs } from '../../components/Breadcrumbs'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { PlusIcon } from '../../components/Icons'
import { Loading, QueryError } from '../../components/PageState'
import { useToast } from '../../components/Toast'
import { useRememberAdminContext } from '../../layouts/AdminLayout'
import { useDocumentTitle } from '../../lib/hooks'
import { countLabel, TASK_FORMS } from '../../lib/text'
import { SubjectDialog } from './dialogs'

type Dialog = { kind: 'edit'; subject?: AdminSubject } | { kind: 'delete'; subject: AdminSubject }

const groupNotFound = { title: 'Группа не найдена', backTo: '/admin/groups', backLabel: '← К группам' }

export function GroupPage() {
  const id = Number(useParams().groupId)
  const details = useAdminGroup(id)
  const remove = useDeleteSubject()
  const toast = useToast()
  const navigate = useNavigate()
  const [dialog, setDialog] = useState<Dialog | null>(null)

  useRememberAdminContext(details.data?.group.id)
  useDocumentTitle(details.data ? `Предметы ${details.data.group.name}` : undefined)

  if (details.isPending) return <Loading />
  if (details.isError) return <QueryError error={details.error} retry={() => details.refetch()} notFound={groupNotFound} />

  const { group, subjects } = details.data

  return (
    <div>
      <Breadcrumbs items={[{ label: 'Группы', to: '/admin/groups' }, { label: group.name }]} />
      <div className="page-head">
        <div className="title-with-tag">
          <h1 className="page-title">Предметы {group.name}</h1>
          {group.hidden && <span className="tag tag-neutral">Группа скрыта</span>}
        </div>
        <button type="button" className="btn btn-primary" onClick={() => setDialog({ kind: 'edit' })}>
          <PlusIcon />
          Предмет
        </button>
      </div>

      {subjects.length === 0 ? (
        <p className="muted">В группе пока нет предметов.</p>
      ) : (
        <div className="table-wrap fi">
          <table className="table">
            <thead>
              <tr>
                <th>Название</th>
                <th>Заданий</th>
                <th>
                  <span className="visually-hidden">Действия</span>
                </th>
              </tr>
            </thead>
            <tbody>
              {subjects.map((subject) => (
                <tr key={subject.id}>
                  <td className="cell-title">
                    <Link to={`/admin/subjects/${subject.id}`}>{subject.name}</Link>
                  </td>
                  <td>{subject.tasksCount}</td>
                  <td>
                    <div className="rowact">
                      <button type="button" className="btn btn-secondary" onClick={() => navigate(`/admin/subjects/${subject.id}`)}>
                        Задания
                      </button>
                      <button type="button" className="btn btn-secondary" onClick={() => setDialog({ kind: 'edit', subject })}>
                        Изменить
                      </button>
                      <button type="button" className="btn btn-ghost" onClick={() => setDialog({ kind: 'delete', subject })}>
                        Удалить
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {dialog?.kind === 'edit' && <SubjectDialog groupId={group.id} subject={dialog.subject} onClose={() => setDialog(null)} />}
      {dialog?.kind === 'delete' && (
        <ConfirmDialog
          title={`Удалить предмет «${dialog.subject.name}»?`}
          message={
            dialog.subject.tasksCount > 0
              ? `Вместе с ним удалятся ${countLabel(dialog.subject.tasksCount, TASK_FORMS)}. Это нельзя отменить.`
              : 'Это нельзя отменить.'
          }
          confirmLabel="Удалить"
          onConfirm={() => remove.mutateAsync(dialog.subject.id).then(() => toast('Предмет удалён'))}
          onClose={() => setDialog(null)}
        />
      )}
    </div>
  )
}
