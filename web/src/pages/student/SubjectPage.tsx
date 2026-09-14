import { useParams } from 'react-router'
import { Breadcrumbs } from '../../components/Breadcrumbs'
import { Corners } from '../../components/Corners'
import { NotFound } from '../../components/PageState'
import { TaskLink } from '../../components/TaskLink'
import { useSelection, useStudentShell } from '../../layouts/StudentLayout'
import { findGroup, findSubject } from '../../lib/catalog'
import { useDocumentTitle } from '../../lib/hooks'
import { countLabel, ordinal, TASK_FORMS } from '../../lib/text'

export function SubjectPage() {
  const { catalog } = useStudentShell()
  const params = useParams()
  const group = findGroup(catalog, Number(params.groupId))
  const subject = group && findSubject(group, Number(params.subjectId))

  useSelection(group?.id, subject?.id)
  useDocumentTitle(subject?.name)

  if (!group || !subject) {
    return <NotFound title="Предмет не найден" message="Возможно, его переименовали или скрыли." />
  }

  return (
    <div>
      <Breadcrumbs
        items={[
          { label: 'Группы', to: '/' },
          { label: group.name, to: `/groups/${group.id}` },
          { label: subject.name },
        ]}
      />
      <h1 className="page-title ri">{subject.name}</h1>
      <p className="subtitle ri delay-1">{countLabel(subject.tasks.length, TASK_FORMS)}</p>

      {subject.tasks.length === 0 ? (
        <div className="blueprint empty">
          <Corners />
          Задания появятся здесь
        </div>
      ) : (
        <div className="grid-2 stg" key={subject.id}>
          {subject.tasks.map((task, index) => (
            <TaskLink key={task.id} task={task} className="card blueprint lift card-link">
              <Corners />
              <div className="card-kicker">Задание {ordinal(index)}</div>
              <div className="card-title">{task.name}</div>
              {task.comment && <div className="card-body">{task.comment}</div>}
            </TaskLink>
          ))}
        </div>
      )}
    </div>
  )
}
