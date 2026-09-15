import { Link, useParams } from 'react-router'
import { usePresence } from '../../api/queries'
import { Corners } from '../../components/Corners'
import { NotFound } from '../../components/PageState'
import { PresenceCard } from '../../components/PresenceCard'
import { useSelection, useStudentShell } from '../../layouts/StudentLayout'
import { findGroup } from '../../lib/catalog'
import { groupsWithLessonFirst, matchLesson } from '../../lib/currentLesson'
import { useDocumentTitle } from '../../lib/hooks'
import { countLabel, ordinal, SUBJECT_FORMS, TASK_FORMS } from '../../lib/text'

export function HomePage() {
  const { catalog } = useStudentShell()
  const params = useParams()
  const presence = usePresence()
  const groupId = params.groupId === undefined ? undefined : Number(params.groupId)
  // Without a chosen group the page shows the first one in the sidebar, which during a lesson is the group having it.
  const lessonNow = presence.data?.status === 'in_class' ? matchLesson(catalog, presence.data.current) : null
  const group = groupId === undefined ? groupsWithLessonFirst(catalog.groups, lessonNow)[0] : findGroup(catalog, groupId)

  useSelection(group?.id, undefined)
  useDocumentTitle(group?.name)

  if (groupId !== undefined && !group) {
    return <NotFound title="Группа не найдена" message="Возможно, её переименовали или скрыли." />
  }

  return (
    <div>
      <div className="kicker ri">Учебные материалы</div>
      <h1 className="page-title page-title--hero ri delay-1">Задания и материалы для студентов</h1>
      <p className="lead ri delay-2">Выберите свою группу, а затем предмет. Внутри — задания и всё, что к ним нужно.</p>

      <PresenceCard />

      {!group ? (
        <div className="blueprint empty">
          <Corners />
          Задания пока не опубликованы
        </div>
      ) : (
        <section aria-labelledby="group-title">
          <div className="section-head">
            <h2 id="group-title">{group.name}</h2>
            <span className="tag tag-outline">{countLabel(group.subjects.length, SUBJECT_FORMS)}</span>
          </div>

          {group.subjects.length === 0 ? (
            <div className="blueprint empty">
              <Corners />
              Предметы появятся здесь
            </div>
          ) : (
            <div className="grid-2 stg" key={group.id}>
              {group.subjects.map((subject, index) => (
                <Link
                  key={subject.id}
                  to={`/groups/${group.id}/subjects/${subject.id}`}
                  className="card blueprint lift card-link"
                >
                  <Corners />
                  <div className="card-kicker">Предмет {ordinal(index)}</div>
                  <div className="card-title">{subject.name}</div>
                  <div className="card-foot">
                    <span>{countLabel(subject.tasks.length, TASK_FORMS)}</span>
                    <span className="arr" aria-hidden="true">
                      →
                    </span>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </section>
      )}
    </div>
  )
}
