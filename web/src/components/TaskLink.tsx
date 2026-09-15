import type { ReactNode } from 'react'
import { Link } from 'react-router'
import type { CatalogTask } from '../api/types'

interface TaskLinkProps {
  task: CatalogTask
  className?: string
  children: ReactNode
}

/**
 * Opens a task the way it can be shown: a document as a page on the site,
 * any other link (a PDF, an external site) in a new tab.
 */
export function TaskLink({ task, className, children }: TaskLinkProps) {
  // A locked task, even a plain link, opens on the site first to ask for the password.
  if (task.isDocument || task.locked) {
    return (
      <Link to={`/tasks/${task.id}`} className={className}>
        {children}
      </Link>
    )
  }
  if (task.href) {
    return (
      <a href={task.href} target="_blank" rel="noopener noreferrer" className={className}>
        {children}
        <span className="visually-hidden"> (откроется в новой вкладке)</span>
      </a>
    )
  }
  return (
    <div className={className} aria-disabled="true">
      {children}
    </div>
  )
}
