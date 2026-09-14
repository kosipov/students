import type { Catalog } from '../../api/types'
import { TaskLink } from '../../components/TaskLink'
import { searchTasks } from '../../lib/catalog'

export function SearchResults({ catalog, query }: { catalog: Catalog; query: string }) {
  const results = searchTasks(catalog, query)

  return (
    <section aria-labelledby="search-title">
      <h1 id="search-title" className="page-title page-title--search ri">
        Результаты поиска
      </h1>
      {results.length === 0 ? (
        <p className="muted" role="status">
          Ничего не нашлось. Попробуйте другое слово.
        </p>
      ) : (
        <div className="stg">
          {results.map(({ task, group, subject }) => (
            <TaskLink key={task.id} task={task} className="result-row rowhover">
              <span>
                <span className="result-name">{task.name}</span>
                <span className="result-where">
                  {group.name} · {subject.name}
                </span>
              </span>
              <span className="arr" aria-hidden="true">
                →
              </span>
            </TaskLink>
          ))}
        </div>
      )}
    </section>
  )
}
