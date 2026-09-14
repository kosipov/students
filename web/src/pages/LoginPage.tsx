import { useState, type FormEvent } from 'react'
import { Link, Navigate, useNavigate, useSearchParams } from 'react-router'
import { ApiError } from '../api/client'
import { useMe, useSignIn } from '../api/queries'
import { Corners } from '../components/Corners'
import { Footer, Header } from '../components/Header'
import { TextField } from '../components/TextField'
import { useToast } from '../components/Toast'
import { useDocumentTitle } from '../lib/hooks'
import { safeNextPath } from '../lib/text'

export function LoginPage() {
  useDocumentTitle('Вход')
  const [searchParams] = useSearchParams()
  const next = safeNextPath(searchParams.get('next'), '/admin')
  const navigate = useNavigate()
  const toast = useToast()
  const me = useMe()
  const signIn = useSignIn()
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')

  if (me.data) {
    return <Navigate to={next} replace />
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    signIn.mutate(
      { login, password },
      {
        onSuccess: () => {
          toast('Вы вошли как администратор')
          navigate(next, { replace: true })
        },
      },
    )
  }

  const error = signIn.error instanceof ApiError ? signIn.error.message : signIn.error ? 'Не удалось войти, попробуйте ещё раз' : ''

  return (
    <div className="shell">
      <Header />
      <main className="login-wrap" id="main">
        <form className="blueprint login-card ri" onSubmit={submit} noValidate>
          <Corners />
          <h1>Вход</h1>
          {error && (
            <p className="form-error" role="alert">
              {error}
            </p>
          )}
          <TextField label="Логин" value={login} onChange={setLogin} autoComplete="username" autoFocus required />
          <TextField
            label="Пароль"
            type="password"
            value={password}
            onChange={setPassword}
            autoComplete="current-password"
            required
          />
          <button type="submit" className="btn btn-primary btn-block" disabled={signIn.isPending}>
            {signIn.isPending ? 'Входим…' : 'Войти'}
          </button>
          <Link className="btn btn-ghost" to="/" style={{ marginTop: 'var(--space-3)' }}>
            ← К заданиям
          </Link>
        </form>
      </main>
      <Footer />
    </div>
  )
}
