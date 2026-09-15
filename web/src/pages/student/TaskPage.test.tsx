import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { Catalog, Task } from '../../api/types'
import { StudentLayout } from '../../layouts/StudentLayout'
import { mockApi, renderWithProviders } from '../../test/render'
import { TaskPage } from './TaskPage'

const catalog: Catalog = { groups: [{ id: 1, name: '1435-ИРо', subjects: [{ id: 10, name: 'Веб', tasks: [] }] }] }

const locked: Task = {
  id: 5,
  name: 'Итоговый тест',
  comment: 'Откроется на паре',
  href: '',
  hidden: false,
  isDocument: false,
  protected: true,
  locked: true,
  group: { id: 1, name: '1435-ИРо' },
  subject: { id: 10, name: 'Веб' },
  contentHtml: null,
  updatedAt: null,
}

const unlocked: Task = { ...locked, locked: false, isDocument: true, href: 'https://1drv.ms/t/c/1/secret', contentHtml: '<h1>Вопросы теста</h1>' }

afterEach(() => vi.unstubAllGlobals())

function renderTask(handlers: Parameters<typeof mockApi>[0]) {
  const fetchMock = mockApi({
    'GET /api/catalog': () => ({ status: 200, body: catalog }),
    'GET /api/auth/me': () => ({ status: 401, body: { error: 'Нужно войти' } }),
    ...handlers,
  })
  renderWithProviders(
    <Routes>
      <Route element={<StudentLayout />}>
        <Route path="tasks/:taskId" element={<TaskPage />} />
      </Route>
    </Routes>,
    { path: '/tasks/5' },
  )
  return fetchMock
}

describe('TaskPage with a password', () => {
  it('asks for the password and shows the task after the right one', async () => {
    let isUnlocked = false
    const fetchMock = renderTask({
      'GET /api/tasks/5': () => ({ status: 200, body: isUnlocked ? unlocked : locked }),
      'POST /api/tasks/5/unlock': (init) => {
        const { password } = JSON.parse(String(init.body)) as { password: string }
        if (password !== 'k3y') return { status: 422, body: { error: 'Неверный пароль', fields: { password: 'Неверный пароль' } } }
        isUnlocked = true
        return { status: 204 }
      },
    })

    expect(await screen.findByRole('heading', { name: 'Задание защищено паролем' })).toBeInTheDocument()
    expect(screen.getByText('Откроется на паре')).toBeInTheDocument()
    expect(screen.queryByRole('link', { name: 'Открыть файл' })).not.toBeInTheDocument()

    const input = screen.getByLabelText('Пароль')
    await userEvent.type(input, 'nope{Enter}')
    expect(await screen.findByText('Неверный пароль')).toBeInTheDocument()

    await userEvent.clear(input)
    await userEvent.type(input, 'k3y{Enter}')

    expect(await screen.findByRole('heading', { name: 'Вопросы теста' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Открыть файл' })).toHaveAttribute('href', 'https://1drv.ms/t/c/1/secret')
    const unlockCalls = fetchMock.mock.calls.filter(([url]) => url === '/api/tasks/5/unlock')
    expect(unlockCalls).toHaveLength(2)
    expect(unlockCalls[0]?.[1]?.headers).toMatchObject({ 'X-Requested-With': 'fetch' })
  })

  it('shows the limit message when there were too many attempts', async () => {
    renderTask({
      'GET /api/tasks/5': () => ({ status: 200, body: locked }),
      'POST /api/tasks/5/unlock': () => ({ status: 429, body: { error: 'Слишком много попыток. Попробуйте через 12 мин.' } }),
    })

    await userEvent.type(await screen.findByLabelText('Пароль'), 'guess{Enter}')
    expect(await screen.findByText('Слишком много попыток. Попробуйте через 12 мин.')).toBeInTheDocument()
  })
})
