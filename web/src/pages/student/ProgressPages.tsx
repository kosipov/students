import { Link, useParams } from 'react-router'
import { useRoster, useStudentProgress } from '../../api/queries'
import type { MarkStatus, StudentProgressSubject } from '../../api/types'
import { Corners } from '../../components/Corners'
import { Loading, QueryError } from '../../components/PageState'
import { useSelection, useStudentShell } from '../../layouts/StudentLayout'
import { useDocumentTitle } from '../../lib/hooks'
import { countLabel, STUDENT_FORMS } from '../../lib/text'

const statusLabel: Record<MarkStatus | '', string> = {
  '': 'Нет отметки',
  in_progress: 'В работе',
  submitted: 'Сдано',
  accepted: 'Принято',
}

const statusTag: Record<MarkStatus | '', string> = {
  '': 'tag-neutral',
  in_progress: 'tag-warn',
  submitted: 'tag-info',
  accepted: 'tag-accent',
}

/** Entry point of the public progress: choose a group. */
export function ProgressGroupsPage() {
  const { catalog } = useStudentShell()
  useSelection(undefined, undefined)
  useDocumentTitle('Успеваемость')

  return (
    <div>
      <div className="kicker ri">Успеваемость</div>
      <h1 className="page-title page-title--hero ri delay-1">Отметки о лабораторных</h1>
      <p className="lead ri delay-2">Выберите свою группу, найдите себя в списке и смотрите отметки по каждому предмету.</p>

      {catalog.groups.length === 0 ? (
        <div className="blueprint empty">
          <Corners />
          Группы пока не опубликованы
        </div>
      ) : (
        <div className="grid-2 stg">
          {catalog.groups.map((group) => (
            <Link key={group.id} to={`/progress/groups/${group.id}`} className="card blueprint lift card-link">
              <Corners />
              <div className="card-kicker">Группа</div>
              <div className="card-title">{group.name}</div>
              <div className="card-foot">
                <span>Список группы</span>
                <span className="arr" aria-hidden="true">
                  →
                </span>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

/** The roster of one group: the student finds their own name. */
export function ProgressRosterPage() {
  const groupId = Number(useParams().groupId)
  const roster = useRoster(groupId)

  useSelection(groupId, undefined)
  useDocumentTitle(roster.data ? `Успеваемость ${roster.data.group.name}` : undefined)

  if (roster.isPending) return <Loading />
  if (roster.isError) {
    return (
      <QueryError
        error={roster.error}
        retry={() => roster.refetch()}
        notFound={{ title: 'Группа не найдена', backTo: '/progress', backLabel: '← К выбору группы' }}
      />
    )
  }

  return (
    <div>
      <div className="kicker ri">Успеваемость</div>
      <h1 className="page-title ri delay-1">{roster.data.group.name}</h1>
      <p className="lead ri delay-2">Найдите себя в списке и нажмите на фамилию.</p>

      {roster.data.students.length === 0 ? (
        <div className="blueprint empty">
          <Corners />
          Список группы пока не добавлен
        </div>
      ) : (
        <section aria-labelledby="roster-heading">
          <div className="section-head">
            <h2 id="roster-heading">Студенты</h2>
            <span className="tag tag-outline">{countLabel(roster.data.students.length, STUDENT_FORMS)}</span>
          </div>
          <div className="table-wrap fi">
            <table className="table">
              <thead>
                <tr>
                  <th>Фамилия и имя</th>
                  <th>
                    <span className="visually-hidden">Перейти</span>
                  </th>
                </tr>
              </thead>
              <tbody>
                {roster.data.students.map((student) => (
                  <tr key={student.id}>
                    <td className="cell-title">
                      <Link to={`/progress/students/${student.id}`}>{student.name}</Link>
                    </td>
                    <td>
                      <div className="rowact">
                        <Link to={`/progress/students/${student.id}`} className="btn btn-secondary">
                          Мои отметки
                        </Link>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}
    </div>
  )
}

function SubjectProgress({ subject }: { subject: StudentProgressSubject }) {
  const accepted = subject.tasks.filter((task) => task.status === 'accepted').length
  return (
    <section aria-labelledby={`subject-${subject.id}`} className="roster">
      <div className="section-head">
        <h2 id={`subject-${subject.id}`}>{subject.name}</h2>
        <span className="tag tag-outline">
          Принято {accepted} из {subject.tasks.length}
        </span>
      </div>
      {subject.tasks.length === 0 ? (
        <p className="muted">Заданий пока нет.</p>
      ) : (
        <div className="table-wrap fi">
          <table className="table">
            <thead>
              <tr>
                <th>Задание</th>
                <th>Отметка</th>
                <th>Комментарий</th>
              </tr>
            </thead>
            <tbody>
              {subject.tasks.map((task) => (
                <tr key={task.id}>
                  <td className="cell-title">{task.name}</td>
                  <td>
                    <span className={`tag ${statusTag[task.status]}`}>{statusLabel[task.status]}</span>
                  </td>
                  <td>{task.markComment || <span className="muted">—</span>}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}

/** The student's own marks on every subject of their group. */
export function StudentProgressPage() {
  const studentId = Number(useParams().studentId)
  const progress = useStudentProgress(studentId)

  useSelection(progress.data?.group.id, undefined)
  useDocumentTitle(progress.data ? progress.data.student.name : undefined)

  if (progress.isPending) return <Loading />
  if (progress.isError) {
    return (
      <QueryError
        error={progress.error}
        retry={() => progress.refetch()}
        notFound={{ title: 'Студент не найден', backTo: '/progress', backLabel: '← К выбору группы' }}
      />
    )
  }

  const { student, group, subjects } = progress.data
  return (
    <div>
      <div className="kicker ri">
        <Link to={`/progress/groups/${group.id}`} className="kicker-link">
          Успеваемость {group.name}
        </Link>
      </div>
      <h1 className="page-title page-title--hero ri delay-1">{student.name}</h1>
      <p className="lead ri delay-2">Отметки о выполнении лабораторных работ.</p>

      {subjects.length === 0 ? (
        <div className="blueprint empty">
          <Corners />
          Предметы появятся здесь
        </div>
      ) : (
        subjects.map((subject) => <SubjectProgress key={subject.id} subject={subject} />)
      )}
    </div>
  )
}
