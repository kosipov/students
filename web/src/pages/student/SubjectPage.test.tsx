import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Route, Routes, useLocation } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { Catalog } from '../../api/types'
import { StudentLayout } from '../../layouts/StudentLayout'
import { mockApi, renderWithProviders } from '../../test/render'
import { SubjectPage } from './SubjectPage'

const catalog: Catalog = {
  groups: [
    {
      id: 1,
      name: '1435-ИРо',
      subjects: [
        {
          id: 10,
          name: 'Веб-разработка',
          tasks: [
            { id: 100, name: 'Темы курсовых', comment: '', href: 'https://example.com/1', isDocument: false, categories: ['Курсовые'] },
            { id: 101, name: 'Методичка', comment: '', href: 'https://example.com/2', isDocument: false, categories: ['Методички', 'Курсовые'] },
            { id: 102, name: 'Лабораторная', comment: '', href: 'https://example.com/3', isDocument: false, categories: [] },
          ],
        },
        { id: 20, name: 'Без категорий', tasks: [{ id: 200, name: 'Задание', comment: '', href: '', isDocument: false, categories: [] }] },
      ],
    },
  ],
}

function Location() {
  const location = useLocation()
  return <output data-testid="location">{decodeURIComponent(location.pathname + location.search)}</output>
}

function renderSubject(path: string) {
  mockApi({
    'GET /api/catalog': () => ({ status: 200, body: catalog }),
    'GET /api/auth/me': () => ({ status: 401, body: { error: 'Нужно войти' } }),
  })
  return renderWithProviders(
    <>
      <Routes>
        <Route element={<StudentLayout />}>
          <Route path="groups/:groupId/subjects/:subjectId" element={<SubjectPage />} />
        </Route>
      </Routes>
      <Location />
    </>,
    { path },
  )
}

const taskNames = () => screen.queryAllByText(/^(Темы курсовых|Методичка|Лабораторная)$/).map((el) => el.textContent)

afterEach(() => vi.unstubAllGlobals())

describe('SubjectPage filters', () => {
  it('filters tasks by category and keeps the choice in the address', async () => {
    renderSubject('/groups/1/subjects/10')
    const filters = within(await screen.findByRole('group', { name: 'Категории заданий' }))

    expect(filters.getAllByRole('button').map((b) => b.textContent)).toEqual(['Все задания', 'Курсовые', 'Методички'])
    expect(filters.getByRole('button', { name: 'Все задания' })).toHaveAttribute('aria-pressed', 'true')
    expect(taskNames()).toEqual(['Темы курсовых', 'Методичка', 'Лабораторная'])

    await userEvent.click(filters.getByRole('button', { name: 'Методички' }))
    expect(taskNames()).toEqual(['Методичка'])
    expect(screen.getByText('Задание 02')).toBeInTheDocument()
    expect(screen.getByTestId('location')).toHaveTextContent('/groups/1/subjects/10?category=Методички')

    await userEvent.click(filters.getByRole('button', { name: 'Все задания' }))
    expect(taskNames()).toHaveLength(3)
    expect(screen.getByTestId('location')).toHaveTextContent(/^\/groups\/1\/subjects\/10$/)
  })

  it('opens with the category from a shared link, ignoring letter case', async () => {
    renderSubject('/groups/1/subjects/10?category=%D0%BA%D1%83%D1%80%D1%81%D0%BE%D0%B2%D1%8B%D0%B5')
    const filters = within(await screen.findByRole('group', { name: 'Категории заданий' }))

    expect(filters.getByRole('button', { name: 'Курсовые' })).toHaveAttribute('aria-pressed', 'true')
    expect(taskNames()).toEqual(['Темы курсовых', 'Методичка'])
  })

  it('shows all tasks for an unknown category and hides filters without categories', async () => {
    renderSubject('/groups/1/subjects/10?category=Удалённая')
    expect(await screen.findByRole('button', { name: 'Все задания' })).toHaveAttribute('aria-pressed', 'true')
    expect(taskNames()).toHaveLength(3)
  })

  it('has no filters when tasks have no categories', async () => {
    renderSubject('/groups/1/subjects/20')
    expect(await screen.findByRole('heading', { name: 'Без категорий' })).toBeInTheDocument()
    expect(screen.queryByRole('group', { name: 'Категории заданий' })).not.toBeInTheDocument()
  })
})
