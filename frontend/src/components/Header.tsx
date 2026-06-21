import { Logo } from './Logo'
import { Button } from './ui'
import { initialOf } from '../lib/format'
import { useAuth } from '../auth/useAuth'

export function Header() {
  const { user, isAuthenticated, openAuth, logout } = useAuth()

  return (
    <header className="glass sticky top-0 z-40 border-x-0 border-t-0">
      <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4 sm:px-6">
        <a href="/" className="rounded-lg" aria-label="SoundFlow home">
          <Logo />
        </a>

        {isAuthenticated && user ? (
          <div className="flex items-center gap-3">
            <span className="flex items-center gap-2">
              <span
                aria-hidden="true"
                className="grid h-7 w-7 place-items-center rounded-full bg-gradient-to-br from-accent to-success text-xs font-bold text-bg"
              >
                {initialOf(user.name)}
              </span>
              <span className="hidden text-sm font-medium text-text sm:inline">{user.name}</span>
            </span>
            <Button variant="subtle" onClick={logout} className="px-2 py-1.5">
              Log out
            </Button>
          </div>
        ) : (
          <div className="flex items-center gap-1.5">
            <Button variant="subtle" onClick={() => openAuth('login')} className="px-3 py-1.5">
              Log in
            </Button>
            <Button variant="ghost" onClick={() => openAuth('signup')} className="px-3 py-1.5">
              Sign up
            </Button>
          </div>
        )}
      </div>
    </header>
  )
}
