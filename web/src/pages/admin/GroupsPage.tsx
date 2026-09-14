import { useState } from 'react'
import { Link } from 'react-router'
import { useAdminGroups, useDeleteGroup, useUpdateGroup } from '../../api/queries'
import type { AdminGroup } from '../../api/types'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { PlusIcon } from '../../components/Icons'
import { Loading, QueryError } from '../../components/PageState'
import { useToast } from '../../components/Toast'
import { useDocumentTitle } from '../../lib/hooks'
import { countLabel, SUBJECT_FORMS, TASK_FORMS } from '../../lib/text'
import { GroupDialog } from './dialogs'

type Dialog = { kind: 'edit'; group?: AdminGroup } | { kind: 'delete'; group: AdminGroup }

export function GroupsPage() {
  useDocumentTitle('Группы')
  const groups = useAdminGroups()
  const update = useUpdateGroup()
  const remove = useDeleteGroup()
  const toast = useToast()
  const [dialog, setDialog] = useState<Dialog | null>(null)

  const toggle = (group: AdminGroup) =>
    update.mutate(
      { id: group.id, hidden: !group.hidden },
      {
        onSuccess: (saved) => toast(saved.hidden ? 'Группа скрыта' : 'Группа видна студентам'),
        onError: () => toast('Не удалось изменить видимость'),
      },
    )

  return (
    <div>
      <div className="page-head">
        <h1 className="page-title">Группы</h1>
        <button type="button" className="btn btn-primary" onClick={() => setDialog({ kind: 'edit' })}>
          <PlusIcon />
          Группа
        </button>
      </div>

      {groups.isPending && <Loading />}
      {groups.isError && <QueryError error={groups.error} retry={() => groups.refetch()} />}
      {groups.data &&
        (groups.data.length === 0 ? (
          <p className="muted">Групп пока нет. Создайте первую.</p>
        ) : (
          <div className="table-wrap fi">
            <table className="table">
              <thead>
                <tr>
                  <th>Название</th>
                  <th>Предметов</th>
                  <th>Заданий</th>
                  <th>Видимость</th>
                  <th>
                    <span className="visually-hidden">Действия</span>
                  </th>
                </tr>
              </thead>
              <tbody>
                {groups.data.map((group) => (
                  <tr key={group.id}>
                    <td className="cell-title">
                      <Link to={`/admin/groups/${group.id}`}>{group.name}</Link>
                    </td>
                    <td>{group.subjectsCount}</td>
                    <td>{group.tasksCount}</td>
                    <td>
                      <span className={`tag ${group.hidden ? 'tag-neutral' : 'tag-accent'}`}>{group.hidden ? 'Скрыта' : 'Видна'}</span>
                    </td>
                    <td>
                      <div className="rowact">
                        <Link className="btn btn-secondary" to={`/admin/groups/${group.id}`}>
                          Предметы
                        </Link>
                        <button type="button" className="btn btn-secondary" onClick={() => toggle(group)}>
                          {group.hidden ? 'Показать' : 'Скрыть'}
                        </button>
                        <button type="button" className="btn btn-secondary" onClick={() => setDialog({ kind: 'edit', group })}>
                          Изменить
                        </button>
                        <button type="button" className="btn btn-ghost" onClick={() => setDialog({ kind: 'delete', group })}>
                          Удалить
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ))}

      {dialog?.kind === 'edit' && <GroupDialog group={dialog.group} onClose={() => setDialog(null)} />}
      {dialog?.kind === 'delete' && (
        <ConfirmDialog
          title={`Удалить группу «${dialog.group.name}»?`}
          message={
            dialog.group.subjectsCount > 0
              ? `Вместе с ней удалятся ${countLabel(dialog.group.subjectsCount, SUBJECT_FORMS)} и ${countLabel(dialog.group.tasksCount, TASK_FORMS)}. Это нельзя отменить.`
              : 'Это нельзя отменить.'
          }
          confirmLabel="Удалить"
          onConfirm={() => remove.mutateAsync(dialog.group.id).then(() => toast('Группа удалена'))}
          onClose={() => setDialog(null)}
        />
      )}
    </div>
  )
}
