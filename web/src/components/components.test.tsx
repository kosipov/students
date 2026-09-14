import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { describe, expect, it, vi } from 'vitest'
import type { CatalogTask } from '../api/types'
import { Modal } from './Modal'
import { TaskLink } from './TaskLink'

const task: CatalogTask = { id: 7, name: 'Темы', comment: '', href: 'https://1drv.ms/t/c/1/abc', isDocument: true, categories: [] }

describe('TaskLink', () => {
  it('opens a document as a page on the site', () => {
    render(
      <MemoryRouter>
        <TaskLink task={task}>Темы</TaskLink>
      </MemoryRouter>,
    )
    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', '/tasks/7')
    expect(link).not.toHaveAttribute('target')
  })

  it('opens other links in a new tab without access to the opener', () => {
    render(
      <MemoryRouter>
        <TaskLink task={{ ...task, isDocument: false, href: 'https://example.com/file.pdf' }}>Методичка</TaskLink>
      </MemoryRouter>,
    )
    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', 'https://example.com/file.pdf')
    expect(link).toHaveAttribute('target', '_blank')
    expect(link).toHaveAttribute('rel', 'noopener noreferrer')
  })

  it('is not a link when there is nothing to open', () => {
    render(
      <MemoryRouter>
        <TaskLink task={{ ...task, isDocument: false, href: '' }}>Без ссылки</TaskLink>
      </MemoryRouter>,
    )
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })
})

describe('Modal', () => {
  it('focuses the first field and closes on Escape', async () => {
    const onClose = vi.fn()
    render(
      <Modal title="Новая группа" onClose={onClose}>
        <input aria-label="Название" />
        <button type="button">Сохранить</button>
      </Modal>,
    )

    expect(screen.getByRole('dialog', { name: 'Новая группа' })).toBeInTheDocument()
    expect(screen.getByLabelText('Название')).toHaveFocus()

    await userEvent.keyboard('{Escape}')
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('keeps focus inside with Tab', async () => {
    render(
      <Modal title="Диалог" onClose={() => {}}>
        <input aria-label="Поле" />
        <button type="button">Готово</button>
      </Modal>,
    )

    await userEvent.tab()
    expect(screen.getByRole('button', { name: 'Готово' })).toHaveFocus()
    await userEvent.tab()
    expect(screen.getByLabelText('Поле')).toHaveFocus()
  })
})
