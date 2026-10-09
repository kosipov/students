import { useMemo, useState } from 'react'
import { useParams } from 'react-router'
import { ApiError } from '../../api/client'
import { useAdminProgress, useDeleteStudent, useImportFromCampus, useSetMark } from '../../api/queries'
import type { MarkStatus, ProgressStudent } from '../../api/types'
import { Breadcrumbs } from '../../components/Breadcrumbs'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { PlusIcon } from '../../components/Icons'
import { Loading, QueryError } from '../../components/PageState'
import { useToast } from '../../components/Toast'
import { useRememberAdminContext } from '../../layouts/AdminLayout'
import { useDocumentTitle } from '../../lib/hooks'
import { countLabel, STUDENT_FORMS } from '../../lib/text'
import { ImportStudentsDialog, StudentDialog } from './progressDialogs'

type Dialog =
  | { kind: 'add' }
  | { kind: 'import' }
  | { kind: 'rename'; student: ProgressStudent }
  | { kind: 'delete'; student: ProgressStudent }

/** Mark statuses in the order a cell cycles through them on click. */
const statusCycle: (MarkStatus | '')[] = ['', 'in_progress', 'submitted', 'accepted']

const statusGlyph: Record<MarkStatus | '', string> = {
  '': '—',
  in_progress: '✎',
  submitted: '➔',
  accepted: '✓',
}

const statusLabel: Record<MarkStatus | '', string> = {
  '': 'Нет отметки',
  in_progress: 'В работе',
  submitted: 'Сдано',
  accepted: 'Принято',
}

function nextStatus(status: MarkStatus | ''): MarkStatus | '' {
  const index = statusCycle.indexOf(status)
  return statusCycle[(index + 1) % statusCycle.length] ?? ''
}

const groupNotFound = { title: 'Группа не найдена', backTo: '/admin/groups', backLabel: '← К группам' }

export function ProgressPage() {
  const groupId = Number(useParams().groupId)
  const progress = useAdminProgress(groupId)
  const setMark = useSetMark(groupId)
  const removeStudent = useDeleteStudent(groupId)
  const importCampus = useImportFromCampus(groupId)
  const toast = useToast()
  const [dialog, setDialog] = useState<Dialog | null>(null)
  const [subjectId, setSubjectId] = useState<number | null>(null)

  useRememberAdminContext(progress.data?.group.id)
  useDocumentTitle(progress.data ? `Успеваемость ${progress.data.group.name}` : undefined)

  const subject = useMemo(() => {
    if (!progress.data || progress.data.subjects.length === 0) return undefined
    return progress.data.subjects.find((s) => s.id === subjectId) ?? progress.data.subjects[0]
  }, [progress.data, subjectId])

  if (progress.isPending) return <Loading />
  if (progress.isError) return <QueryError error={progress.error} retry={() => progress.refetch()} notFound={groupNotFound} />

  const { group, subjects, students } = progress.data
  const marksByStudentTask = new Map(students.map((s) => [s.id, new Map(s.marks.map((m) => [m.taskId, m]))]))

  const cycleMark = (studentId: number, taskId: number) => {
    const current = marksByStudentTask.get(studentId)?.get(taskId)?.status ?? ''
    setMark.mutate(
      { studentId, taskId, status: nextStatus(current) },
      { onError: () => toast('Не удалось сохранить отметку, попробуйте ещё раз') },
    )
  }

  return (
    <div>
      <Breadcrumbs
        items={[{ label: 'Группы', to: '/admin/groups' }, { label: group.name, to: `/admin/groups/${group.id}` }, { label: 'Успеваемость' }]}
      />
      <div className="page-head">
        <div className="title-with-tag">
          <h1 className="page-title">Успеваемость {group.name}</h1>
          {group.hidden && <span className="tag tag-neutral">Группа скрыта</span>}
        </div>
        <div className="rowact">
          <button
            type="button"
            className="btn btn-secondary"
            disabled={importCampus.isPending}
            title={`Загрузить список группы «${group.name}» с campus.syktsu.ru`}
            onClick={() =>
              importCampus.mutate(undefined, {
                onSuccess: (result) =>
                  toast(
                    result.skipped > 0
                      ? `В кампусе найдено: ${result.found}, добавлено: ${result.added}, уже были: ${result.skipped}`
                      : `В кампусе найдено: ${result.found}, добавлено: ${result.added}`,
                  ),
                onError: (error) => toast(error instanceof ApiError ? error.message : 'Не удалось получить список из кампуса'),
              })
            }
          >
            {importCampus.isPending ? 'Загружаем…' : 'Из кампуса'}
          </button>
          <button type="button" className="btn btn-secondary" onClick={() => setDialog({ kind: 'import' })}>
            Импорт списка
          </button>
          <button type="button" className="btn btn-primary" onClick={() => setDialog({ kind: 'add' })}>
            <PlusIcon />
            Студент
          </button>
        </div>
      </div>

      {subjects.length === 0 ? (
        <p className="muted">В группе пока нет предметов — сначала добавьте предмет и задания.</p>
      ) : students.length === 0 ? (
        <p className="muted">
          Список группы пуст. Добавьте студентов по одному или импортируйте весь список разом — отметки появятся здесь.
        </p>
      ) : (
        <>
          <div className="chips" role="tablist" aria-label="Предмет">
            {subjects.map((s) => (
              <button
                key={s.id}
                type="button"
                role="tab"
                aria-selected={subject?.id === s.id}
                className={`btn ${subject?.id === s.id ? 'btn-primary' : 'btn-secondary'}`}
                onClick={() => setSubjectId(s.id)}
              >
                {s.name}
              </button>
            ))}
          </div>

          {subject && subject.tasks.length === 0 ? (
            <p className="muted">В предмете «{subject.name}» пока нет заданий.</p>
          ) : (
            subject && (
              <div className="table-wrap fi">
                <table className="table matrix">
                  <thead>
                    <tr>
                      <th className="matrix-student">Студент</th>
                      {subject.tasks.map((task) => (
                        <th key={task.id} className="matrix-task" title={task.hidden ? `${task.name} (скрыто)` : task.name}>
                          {task.name}
                          {task.hidden && <span className="tag tag-neutral">скрыто</span>}
                        </th>
                      ))}
                      <th className="matrix-task" title="Принято заданий">
                        ✓
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {students.map((student) => {
                      const marks = marksByStudentTask.get(student.id)
                      const accepted = subject.tasks.filter((t) => marks?.get(t.id)?.status === 'accepted').length
                      return (
                        <tr key={student.id}>
                          <td className="matrix-student cell-title">{student.name}</td>
                          {subject.tasks.map((task) => {
                            const status = marks?.get(task.id)?.status ?? ''
                            return (
                              <td key={task.id}>
                                <button
                                  type="button"
                                  className={`mark ${status ? `mark--${status}` : ''}`}
                                  title={`${student.name} — ${task.name}: ${statusLabel[status]}. Нажмите, чтобы изменить`}
                                  aria-label={`${student.name}, ${task.name}: ${statusLabel[status]}`}
                                  onClick={() => cycleMark(student.id, task.id)}
                                >
                                  {statusGlyph[status]}
                                </button>
                              </td>
                            )
                          })}
                          <td className="matrix-total">
                            {accepted}/{subject.tasks.length}
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            )
          )}

          <ul className="legend" aria-label="Обозначения">
            {(['in_progress', 'submitted', 'accepted'] as MarkStatus[]).map((status) => (
              <li key={status}>
                <span className={`mark mark--${status}`} aria-hidden="true">
                  {statusGlyph[status]}
                </span>
                {statusLabel[status]}
              </li>
            ))}
            <li>Нажатие на ячейку меняет отметку по кругу</li>
          </ul>

          <section aria-labelledby="roster-title" className="roster">
            <div className="section-head">
              <h2 id="roster-title">Список группы</h2>
              <span className="tag tag-outline">{countLabel(students.length, STUDENT_FORMS)}</span>
            </div>
            <div className="table-wrap fi">
              <table className="table">
                <thead>
                  <tr>
                    <th>Фамилия и имя</th>
                    <th>
                      <span className="visually-hidden">Действия</span>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {students.map((student) => (
                    <tr key={student.id}>
                      <td className="cell-title">{student.name}</td>
                      <td>
                        <div className="rowact">
                          <button type="button" className="btn btn-secondary" onClick={() => setDialog({ kind: 'rename', student })}>
                            Переименовать
                          </button>
                          <button type="button" className="btn btn-ghost" onClick={() => setDialog({ kind: 'delete', student })}>
                            Удалить
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        </>
      )}

      {dialog?.kind === 'add' && <StudentDialog groupId={groupId} onClose={() => setDialog(null)} />}
      {dialog?.kind === 'import' && <ImportStudentsDialog groupId={groupId} onClose={() => setDialog(null)} />}
      {dialog?.kind === 'rename' && <StudentDialog groupId={groupId} student={dialog.student} onClose={() => setDialog(null)} />}
      {dialog?.kind === 'delete' && (
        <ConfirmDialog
          title={`Удалить ${dialog.student.name}?`}
          message="Вместе со студентом удалятся все его отметки. Это нельзя отменить."
          confirmLabel="Удалить"
          onConfirm={() => removeStudent.mutateAsync(dialog.student.id).then(() => toast('Студент удалён'))}
          onClose={() => setDialog(null)}
        />
      )}
    </div>
  )
}
