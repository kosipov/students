import { Link } from 'react-router'
import { ApiError } from '../api/client'
import { useDocumentTitle } from '../lib/hooks'

export function Loading() {
  return <p className="loading">Загружаем…</p>
}

interface NotFoundProps {
  title?: string
  message?: string
  backTo?: string
  backLabel?: string
}

export function NotFound({
  title = 'Страница не найдена',
  message = 'Возможно, её удалили или скрыли.',
  backTo = '/',
  backLabel = '← К заданиям',
}: NotFoundProps) {
  useDocumentTitle(title)
  return (
    <div>
      <h1 className="page-title ri">{title}</h1>
      <p className="subtitle ri delay-1">{message}</p>
      <Link className="btn btn-secondary" to={backTo}>
        {backLabel}
      </Link>
    </div>
  )
}

/** Shows a failed query: 404 as "not found", anything else with a retry button. */
export function QueryError({ error, retry, notFound }: { error: unknown; retry: () => void; notFound?: NotFoundProps }) {
  if (error instanceof ApiError && error.status === 404) {
    return <NotFound {...notFound} />
  }
  return (
    <div>
      <p className="form-error" role="alert">
        {error instanceof ApiError ? error.message : 'Не удалось загрузить данные'}
      </p>
      <button type="button" className="btn btn-secondary" onClick={retry}>
        Попробовать ещё раз
      </button>
    </div>
  )
}
