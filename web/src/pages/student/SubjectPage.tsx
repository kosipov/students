import { useParams, useSearchParams } from 'react-router'
import { Breadcrumbs } from '../../components/Breadcrumbs'
import { Corners } from '../../components/Corners'
import { NotFound } from '../../components/PageState'
import { LockIcon } from '../../components/Icons'
import { TaskLink } from '../../components/TaskLink'
import { useSelection, useStudentShell } from '../../layouts/StudentLayout'
import { categoriesOf, categoryKey, filterByCategory, findGroup, findSubject } from '../../lib/catalog'
import { useDocumentTitle } from '../../lib/hooks'
import { countLabel, ordinal, TASK_FORMS } from '../../lib/text'

export function SubjectPage() {
  const { catalog } = useStudentShell()
  const params = useParams()
  const [searchParams, setSearchParams] = useSearchParams()
  const group = findGroup(catalog, Number(params.groupId))
  const subject = group && findSubject(group, Number(params.subjectId))

  useSelection(group?.id, subject?.id)
  useDocumentTitle(subject?.name)

  if (!group || !subject) {
    return <NotFound title="Предмет не найден" message="Возможно, его переименовали или скрыли." />
  }

  const categories = categoriesOf(subject.tasks)
  // The chosen category lives in the address, so a filtered list can be shared and "back" works.
  const requested = searchParams.get('category')
  const activeCategory = categories.find((name) => requested !== null && categoryKey(name) === categoryKey(requested)) ?? null
  const tasks = filterByCategory(subject.tasks, activeCategory)

  const choose = (category: string | null) =>
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        if (category) next.set('category', category)
        else next.delete('category')
        return next
      },
      { preventScrollReset: true },
    )

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

      {categories.length > 0 && (
        <div className="filters ri delay-2" role="group" aria-label="Категории заданий">
          <button
            type="button"
            className={`btn ${activeCategory === null ? 'btn-primary' : 'btn-secondary'}`}
            aria-pressed={activeCategory === null}
            onClick={() => choose(null)}
          >
            Все задания
          </button>
          {categories.map((category) => (
            <button
              key={category}
              type="button"
              className={`btn ${category === activeCategory ? 'btn-primary' : 'btn-secondary'}`}
              aria-pressed={category === activeCategory}
              onClick={() => choose(category)}
            >
              {category}
            </button>
          ))}
        </div>
      )}

      {subject.tasks.length === 0 ? (
        <div className="blueprint empty">
          <Corners />
          Задания появятся здесь
        </div>
      ) : (
        <div className="grid-2 stg" key={`${subject.id}-${activeCategory ?? ''}`}>
          {tasks.map((task) => (
            <TaskLink key={task.id} task={task} className="card blueprint lift card-link">
              <Corners />
              {/* The number stays the task's place in the full list, so it doesn't change with the filter. */}
              <div className="card-kicker card-kicker-row">
                Задание {ordinal(subject.tasks.indexOf(task))}
                {task.locked && <LockIcon label="Защищено паролем" />}
              </div>
              <div className="card-title">{task.name}</div>
              {task.comment && <div className="card-body">{task.comment}</div>}
            </TaskLink>
          ))}
        </div>
      )}
    </div>
  )
}
