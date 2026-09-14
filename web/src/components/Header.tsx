import { Link, useLocation, useNavigate } from 'react-router'
import { useMe, useSignOut } from '../api/queries'
import { SearchIcon } from './Icons'
import { useToast } from './Toast'

interface HeaderProps {
  search?: { value: string; onChange: (value: string) => void }
  isAdminArea?: boolean
}

export function Header({ search, isAdminArea = false }: HeaderProps) {
  const { data: me } = useMe()
  const signOut = useSignOut()
  const navigate = useNavigate()
  const location = useLocation()
  const toast = useToast()

  const logout = () =>
    signOut.mutate(undefined, {
      onSuccess: () => {
        navigate('/')
        toast('Вы вышли')
      },
      onError: () => toast('Не удалось выйти, попробуйте ещё раз'),
    })

  return (
    <header className="topbar">
      <Link to="/" className="brand" aria-label="Студентам — на главную">
        <span className="brand-mark" aria-hidden="true">
          С
        </span>
        <span className="brand-name">Студентам</span>
        {me && <span className="tag tag-accent">админка</span>}
      </Link>
      <div className="topbar-actions">
        {search && (
          <div className="searchbox" role="search">
            <SearchIcon />
            <label htmlFor="task-search" className="visually-hidden">
              Найти задание
            </label>
            <input
              id="task-search"
              className="input"
              type="search"
              placeholder="Найти задание"
              value={search.value}
              onChange={(event) => search.onChange(event.target.value)}
            />
          </div>
        )}
        {me ? (
          <>
            {!isAdminArea && (
              <Link className="btn btn-secondary" to="/admin">
                Управление
              </Link>
            )}
            <button type="button" className="btn btn-secondary" onClick={logout} disabled={signOut.isPending}>
              Выйти
            </button>
          </>
        ) : (
          me === null &&
          location.pathname !== '/login' && (
            <Link className="btn btn-primary" to="/login">
              Войти
            </Link>
          )
        )}
      </div>
    </header>
  )
}

export function Footer() {
  return (
    <footer className="footer">
      <span>student.kosipov.ru</span>
    </footer>
  )
}
