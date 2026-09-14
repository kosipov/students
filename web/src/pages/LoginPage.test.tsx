import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { mockApi, renderWithProviders } from '../test/render'
import { LoginPage } from './LoginPage'

afterEach(() => vi.unstubAllGlobals())

function renderLogin(path = '/login') {
  return renderWithProviders(
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/admin/*" element={<p>Админка открыта</p>} />
    </Routes>,
    { path },
  )
}

describe('LoginPage', () => {
  it('shows the server message for wrong credentials', async () => {
    mockApi({
      'GET /api/auth/me': () => ({ status: 401, body: { error: 'Нужно войти' } }),
      'POST /api/auth/sign-in': () => ({ status: 401, body: { error: 'Неверный логин или пароль' } }),
    })
    renderLogin()

    await userEvent.type(screen.getByLabelText('Логин'), 'admin')
    await userEvent.type(screen.getByLabelText('Пароль'), 'wrong')
    await userEvent.click(screen.getByRole('button', { name: 'Войти' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('Неверный логин или пароль')
  })

  it('signs in and goes to the requested admin page', async () => {
    const fetchMock = mockApi({
      'GET /api/auth/me': () => ({ status: 401, body: { error: 'Нужно войти' } }),
      'POST /api/auth/sign-in': () => ({ status: 200, body: { userName: 'admin' } }),
    })
    renderLogin('/login?next=%2Fadmin%2Fgroups')

    await userEvent.type(screen.getByLabelText('Логин'), 'admin')
    await userEvent.type(screen.getByLabelText('Пароль'), 'S3cret-pass{Enter}')

    expect(await screen.findByText('Админка открыта')).toBeInTheDocument()
    const signIn = fetchMock.mock.calls.find(([url]) => url === '/api/auth/sign-in')
    expect(signIn?.[1]?.body).toBe(JSON.stringify({ login: 'admin', password: 'S3cret-pass' }))
  })
})
